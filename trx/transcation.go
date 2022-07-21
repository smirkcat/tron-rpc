package trx

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
	"tron/api"
	"tron/core"
	"tron/hexutil"
	"tron/log"
	"tron/service"

	wallet "tron/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/smirkcat/hdwallet"
	"google.golang.org/protobuf/proto"
)

// 每次最多100 个
func getBlockWithHeights(start, end int64) {
	if end-start < 1 {
		return
	}
	var node *service.GrpcClient
againblock:
	if node != nil {
		node.Conn.Close()
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	node = getRandOneNode()
	block, err := node.GetBlockByLimitNext(start, end)
	if err != nil {
		// rpc error: code = DeadlineExceeded desc = context deadline exceeded will get again
		log.Warnf("node get bolck start %d end %d GetBlockByLimitNext err: %v will get again", start, end, err)
		time.Sleep(time.Second * 5)
		goto againblock
	}
	log.Infof("node get bolck start %d end %d length %d", start, end, len(block.Block))
	if len(block.Block) < 1 {
		log.Warnf("get bolck zero lenghth of block start %d end %d, will get again", start, end)
		time.Sleep(time.Second * 5)
		goto againblock
	}
	processBlocks(block)
	node.Conn.Close()
}

func getBlockWithHeight(num int64) error {
	node := getRandOneNode()
	defer node.Conn.Close()
	block, err := node.GetBlockByNum(num)
	if err != nil {
		return err
	}
	processBlock(block)
	return nil
}

func processBlocks(blocks *api.BlockListExtention) {
	for _, v := range blocks.Block {
		processBlock(v)
	}
}

func getTransactionInfo(node *service.GrpcClient, txid string) *core.TransactionInfo {
againtransinfo:
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	transinfo, err := node.GetTransactionInfoById(txid)
	if err != nil {
		// rpc error: code = DeadlineExceeded desc = context deadline exceeded will get again
		log.Warnf("node get txid %s GetTransactionInfoById  err: %v will get again", txid, err)
		time.Sleep(time.Second * 5)
		goto againtransinfo
	}
	return transinfo
}

// 通过translog判断合约转账 如果有转账有扣除，则需调用此方法更精确
var transferid = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

func processBlock(block *api.BlockExtention) {
	height := block.GetBlockHeader().GetRawData().GetNumber()
	node := getRandOneNode()
	defer node.Conn.Close()
	for _, v := range block.Transactions {
		// transaction.ret.contractRe
		txid := hexutil.Encode(v.Txid)
		// https://tronscan.org/#/transaction/fede1aa9e5c5d7bd179fd62e23bdd11e3c1edd0ca51e41070e34a026d6a42569
		if v.Result == nil || !v.Result.Result {
			continue
		}
		rets := v.Transaction.Ret
		if len(rets) < 1 || rets[0].ContractRet != core.Transaction_Result_SUCCESS {
			continue
		}

		log.Debugf("process block height %d txid %s", height, txid)
		var transinfo *core.TransactionInfo
		var fee int64
		// 这里只能有一个
		for _, v1 := range v.Transaction.RawData.Contract {
			if v1.Type == core.Transaction_Contract_TransferContract { //转账合约
				// trx 转账
				unObj := &core.TransferContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}
				form := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
				to := hdwallet.EncodeCheck(unObj.GetToAddress())
				processTransaction(node, Trx, txid, form, to, height, unObj.GetAmount(), fee)
			} else if v1.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
				// trc20 转账
				unObj := &core.TriggerSmartContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}

				contract := hdwallet.EncodeCheck(unObj.GetContractAddress())
				if !IsContract(contract) {
					continue
				}
				from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
				data := unObj.GetData()
				// unObj.Data  https://goethereumbook.org/en/transfer-tokens/ 参考eth 操作
				// 只处理 transfer函数产生的交易
				_, _, flag := processTransferData(data, from)
				if flag { // 只有调用了 transfer(address,uint256) 才处理转账
					// 手续费处理 eth 类似 recipt
					if transinfo == nil {
						transinfo = getTransactionInfo(node, txid)
					}
					if transinfo == nil {
						continue
					}
					fee = transinfo.GetFee()
					// 处理 evenlog 合约转账，如有些合约发起转账并不是全部到账
					// https://tronscan.org/#/address/TWsZk6fs7UisoJAFXiMDXk9aF4PPRzVywZ/transfers
					// https://tronscan.org/#/transaction/0384391ab3ecdf70ffa6e20244718a06b998b8af8a226cda46871dec60b5f14d
					for _, evenlog := range transinfo.Log {
						contract, from, to, amount, flag := processEvenlogData(evenlog)
						if flag {
							processTransaction(node, contract, txid, from, to, height, amount, fee)
						}
					}
				}
			} else if v1.Type == core.Transaction_Contract_TransferAssetContract { //通证转账合约
				// trc10 转账
				unObj := &core.TransferAssetContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}
				contract := hdwallet.EncodeCheck(unObj.GetAssetName())
				form := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
				to := hdwallet.EncodeCheck(unObj.GetToAddress())
				processTransaction(node, contract, txid, form, to, height, unObj.GetAmount(), fee)
			}
		}
	}
}

// 处理合约事件参数
func processEvenlogData(evenlog *core.TransactionInfo_Log) (contract, from, to string, amount int64, flag bool) {
	tmpaddr := evenlog.GetAddress()
	tmpaddr = append([]byte{0x41}, tmpaddr...)
	contract = hdwallet.EncodeCheck(tmpaddr[:])
	if !IsContract(contract) {
		return
	}

	amount = new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Data)).Int64()

	if len(evenlog.Topics) != 3 {
		return
	}
	if transferid != hexutil.Encode(evenlog.Topics[0]) {
		return
	}
	if len(evenlog.Topics[1]) != 32 || len(evenlog.Topics[2]) != 32 {
		return
	}
	evenlog.Topics[1][11] = 0x41
	evenlog.Topics[2][11] = 0x41
	from = hdwallet.EncodeCheck(evenlog.Topics[1][11:])
	to = hdwallet.EncodeCheck(evenlog.Topics[2][11:])
	flag = true
	return
}

// 这个结构目前没有用到 只是记录Trc20合约调用对应转换结果
var mapFunctionTcc20 = map[string]string{
	"a9059cbb": "transfer(address,uint256)",
	"70a08231": "balanceOf(address)",
}

// a9059cbb 4 8
// 00000000000000000000004173d5888eedd05efeda5bca710982d9c13b975f98 32 64
// 0000000000000000000000000000000000000000000000000000000000989680 32 64

// 处理合约参数
func processTransferData(trc20 []byte, from string) (to string, amount int64, flag bool) {
	if len(trc20) >= 68 {
		if hexutil.Encode(trc20[:4]) != "a9059cbb" {
			return
		}
		// 多1位41
		trc20[15] = 65 // 0x41
		to = hdwallet.EncodeCheck(trc20[15:36])
		amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[36:68])).Int64()
		// 不在地址范围内 不处理
		if to == mainAddr || from == mainAddr {
			flag = true
			return
		}
		ac, _ := SearchAccount(to)
		if ac != nil {
			flag = true
		}
		// 归集
		ac, _ = SearchAccount(from)
		if ac != nil {
			flag = true
		}
	}
	return
}

// 处理合约转账参数
func processTransferParameter(to string, amount int64) (data []byte) {
	methodID, _ := hexutil.Decode("a9059cbb")
	addr, _ := hdwallet.DecodeCheck(to)
	paddedAddress := common.LeftPadBytes(addr[1:], 32)
	amountBig := new(big.Int).SetInt64(amount)
	paddedAmount := common.LeftPadBytes(amountBig.Bytes(), 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	return
}

// 处理合约获取余额
func processBalanceOfData(trc20 []byte) (amount int64) {
	if len(trc20) >= 32 {
		amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[0:32])).Int64()
	}
	return
}

// 处理合约获取余额参数
func processBalanceOfParameter(addr string) (data []byte) {
	methodID, _ := hexutil.Decode("70a08231")
	add, _ := hdwallet.DecodeCheck(addr)
	paddedAddress := common.LeftPadBytes(add[1:], 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	return
}

func processTransaction(node *service.GrpcClient, contract, txid, from, to string, blockheight, amount, fee int64) {

	// 合约是否存在
	if !IsContract(contract) {
		return
	}
	// fmt.Printf("contract %s txid %s from %s to %s, blockheight %d amount %d \n",
	// 	contract, txid, from, to, blockheight, amount)
	var types string
	var pub string
	if from == mainAddr { // 提币 or 中转
		ac, err := SearchAccount(to)
		if err != nil {
			log.Error(err)
		}
		if ac != nil {
			pub = ac.PublicKey
			types = Collect // 手续费划转
		} else {
			types = Send
		}
	} else if to == mainAddr { // 归集记录
		ac, err := SearchAccount(from)
		if err != nil {
			log.Error(err)
		}
		if ac != nil {
			types = Collect
		} else {
			types = ReceiveOther
		}
	} else {
		acf, err := SearchAccount(from)
		if err != nil {
			log.Error(err)
		}
		act, err := SearchAccount(to)
		if err != nil {
			log.Error(err)
		}
		if act != nil { // 收币地址
			pub = act.PublicKey
			if acf != nil {
				types = CollectOwn // 站内转账 暂时不可能触发
			} else {
				types = Receive
				go collectall(to) // 归集检测
			}
		} else {
			if acf != nil {
				types = CollectSend // 转账到外面地址 异常
			} else {
				return // 不处理 都不是平台的地址
			}
		}
	}

	if fee == 0 {
		transinfo, err := node.GetTransactionInfoById(txid)
		if err != nil {
			log.Error(err)
		} else {
			fee = transinfo.GetFee()
		}
	}

	_, decimalnum := chargeContract(contract)
	var trans = &Transactions{
		TxID:        txid,
		Contract:    contract,
		Type:        types,
		BlockHeight: blockheight,
		Amount:      decimal.New(amount, -decimalnum).String(),
		Fee:         decimal.New(fee, -trxdecimal).String(),
		Timestamp:   time.Now().Unix(),
		PublicKey:   pub,
		Address:     to,
		FromAddress: from,
	}

	_, err := dbengine.InsertTransactions(trans)
	if err != nil {
		log.Infof("InsertTransactions %v err %v ", trans, err)
	}
}

// 转账合约燃烧 trx数量 单位 sun 默认5trx 转账一笔大概消耗能量 0.26trx
var feelimit int64 = 5000000

// 转币
func send(key *ecdsa.PrivateKey, contract, to string, amount decimal.Decimal) (string, error) {
	node := getRandOneNode()
	defer node.Conn.Close()
	typs, decimalnum := chargeContract(contract)
	var amountdecimal = decimal.New(1, decimalnum)
	amountac, _ := amount.Mul(amountdecimal).Float64()
	switch typs {
	case Trc10:
		return node.TransferAsset(key, contract, to, int64(amountac))
	case Trx:
		return node.Transfer(key, to, int64(amountac))
	case Trc20:
		data := processTransferParameter(to, int64(amountac))
		return node.TransferContract(key, contract, data, feelimit)
	}
	return "", fmt.Errorf("the type %s not support now", typs)
}

// 往外转 提币
func sendOut(contract, to string, amount decimal.Decimal) (string, error) {
	return send(mainAccout, contract, to, amount)
}

// 往地址转手续费
func sendFee(to string, amount decimal.Decimal) (string, error) {
	return send(mainAccout, Trx, to, amount)
}

// 归集
func sendIn(contract, from string, amount decimal.Decimal) (string, error) {
	var accout *ecdsa.PrivateKey
	accout, err := loadAccount(from)
	if err != nil {
		return "", err
	}
	return send(accout, contract, mainAddr, amount)
}

// 交易记录
func recentTransactions(contract, addr string, count, skip int) ([]wallet.Transactions, error) {
	re, err := dbengine.GetTransactions(contract, addr, count, skip)
	lens := len(re)
	ral := make([]wallet.Transactions, lens)
	if err != nil {
		return ral, err
	}
	var account = "go-tron-" + contract + "-walletrpc"
	for i := 0; i < lens; i++ {
		ral[i].Address = re[i].Address
		ral[i].FromAddress = re[i].FromAddress
		ral[i].PublicKey = re[i].PublicKey
		ral[i].Fee = json.Number(re[i].Fee)
		ral[i].Amount = json.Number(re[i].Amount)
		ral[i].Category = re[i].Type
		ral[i].Confirmations = blockHeightTop - re[i].BlockHeight + 1
		ral[i].Time = re[i].Timestamp
		ral[i].TimeReceived = re[i].Timestamp
		ral[i].TxID = re[i].TxID
		ral[i].BlockIndex = re[i].BlockHeight
		ral[i].Account = account
	}
	return ral, nil
}

// 归集记录
func collectTransactions(contract string, sTime, eTime int64) ([]wallet.SummaryData, error) {
	re, err := dbengine.GetCollestTransactions(sTime, eTime, contract)
	lens := len(re)
	ral := make([]wallet.SummaryData, lens)
	if err != nil {
		return ral, err
	}
	var account = "go-tron-" + contract + "-walletrpc"
	for i := 0; i < lens; i++ {
		ral[i].Address = re[i].Address
		ral[i].PublicKey = re[i].PublicKey
		ral[i].FromAddress = re[i].FromAddress
		ral[i].Fee = re[i].Fee
		ral[i].Amount = re[i].Amount
		ral[i].Category = re[i].Type
		ral[i].Time = re[i].Timestamp
		ral[i].TimeReceived = re[i].Timestamp
		ral[i].Blocktime = re[i].Timestamp
		ral[i].TxID = re[i].TxID
		ral[i].BlockIndex = re[i].BlockHeight
		ral[i].Account = account
	}
	return ral, nil
}
