package trx

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
	"tron/api"
	"tron/common/base58"
	"tron/common/hexutil"
	"tron/core"
	"tron/log"

	wallet "tron/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/shopspring/decimal"
)

// 每次最多100 个
func getBlockWithHeights(start, end int64) error {
	node := getRandOneNode()
	defer node.Conn.Close()
	if end-start < 1 {
		return nil
	}
againblock:
	block, err := node.GetBlockByLimitNext(start, end)
	if err != nil {
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
	return nil
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

func processBlock(block *api.BlockExtention) {
	height := block.GetBlockHeader().GetRawData().GetNumber()
	node := getRandOneNode()
	defer node.Conn.Close()
	for _, v := range block.Transactions {
		txid := hexutil.Encode(v.Txid)
		for _, v1 := range v.Transaction.RawData.Contract {
			if v1.Type == core.Transaction_Contract_TransferContract { //转账合约
				// trx 转账
				unObj := &core.TransferContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}
				form := base58.EncodeCheck(unObj.GetOwnerAddress())
				to := base58.EncodeCheck(unObj.GetToAddress())
				processTransaction(Trx, txid, form, to, height, unObj.GetAmount(), 0)
				// break
			} else if v1.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
				// trc20 转账
				unObj := &core.TriggerSmartContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}
				contract := base58.EncodeCheck(unObj.GetContractAddress())
				form := base58.EncodeCheck(unObj.GetOwnerAddress())
				data := unObj.GetData()
				// unObj.Data  https://goethereumbook.org/en/transfer-tokens/ 参考eth 操作
				to, amount, flag := processTransferData(data)
				if flag { // 只有调用了 transfer(address,uint256) 才是转账
					processTransaction(contract, txid, form, to, height, amount, 0)
				}
				// break
			} else if v1.Type == core.Transaction_Contract_TransferAssetContract { //通证转账合约
				// trc10 转账
				unObj := &core.TransferAssetContract{}
				err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
				if err != nil {
					log.Errorf("parse Contract %v err: %v", v1, err)
					continue
				}
				contract := base58.EncodeCheck(unObj.GetAssetName())
				form := base58.EncodeCheck(unObj.GetOwnerAddress())
				to := base58.EncodeCheck(unObj.GetToAddress())
				processTransaction(contract, txid, form, to, height, unObj.GetAmount(), 0)
				// break
			}
		}
	}
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
func processTransferData(trc20 []byte) (to string, amount int64, flag bool) {
	if len(trc20) >= 68 {
		if hexutil.Encode(trc20[:4]) != "a9059cbb" {
			return
		}
		to = base58.EncodeCheck(common.TrimLeftZeroes(trc20[4:36]))
		amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[36:68])).Int64()
		flag = true
	}
	return
}

// 处理合约转账参数
func processTransferParameter(to string, amount int64) (data []byte) {
	methodID, _ := hexutil.Decode("a9059cbb")
	addr, _ := base58.DecodeCheck(to)
	paddedAddress := common.LeftPadBytes(addr, 32)
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
	add, _ := base58.DecodeCheck(addr)
	paddedAddress := common.LeftPadBytes(add, 32)
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	return
}

func processTransaction(contract, txid, from, to string, blockheight, amount, fee int64) {
	// fmt.Printf("contract %s txid %s from %s to %s, blockheight %d amount %d fee %d\n",
	// contract, txid, from, to, blockheight, amount, fee)
	// 合约是否存在
	if !IsContract(contract) {
		return
	}
	var types string
	if from == mainAddr { // 提币 or 中转
		ac, _ := dbengine.SearchAccount(to)
		if ac != nil {
			types = Collect
			// 目前这种情况不会发生
		} else {
			types = Send
		}
	} else if to == mainAddr {
		ac, _ := dbengine.SearchAccount(from)
		if ac != nil {
			types = Collect
			collect(contract, from) // 归集检测
		} else {
			types = ReceiveOther
		}
	} else {
		acf, _ := dbengine.SearchAccount(from)
		act, _ := dbengine.SearchAccount(to)
		if act != nil { // 收币地址
			if acf != nil {
				types = CollectOwn // 站内转账 暂时不可能触发
			} else {
				types = Receive
			}
		} else {
			if acf != nil {
				types = CollectSend // 转账到外面地址 异常
			} else {
				return // 不处理 都不是平台的地址
			}
		}
	}
	var trans = &Transactions{
		TxID:        txid,
		Contract:    contract,
		Type:        types,
		BlockHeight: blockheight,
		Amount:      decimal.New(amount, -6).String(),
		Fee:         decimal.New(fee, -6).String(),
		Timestamp:   time.Now().Unix(),
		Address:     to,
		FromAddress: from,
	}

	_, err := dbengine.InsertTransactions(trans)
	log.Infof("InsertTransactions %v err %v ", trans, err)
}

var num60 = decimal.New(1, 6)

// 转币
func send(key *ecdsa.PrivateKey, contract, to string, amount decimal.Decimal) (string, error) {
	node := getRandOneNode()
	defer node.Conn.Close()
	amount6, _ := amount.Mul(num60).Float64()
	typs := chargeContract(contract)
	switch typs {
	case Trc10:
		return node.TransferAsset(key, contract, to, int64(amount6))
	case Trx:
		return node.Transfer(key, to, int64(amount6))
	case Trc20:
		data := processTransferParameter(to, int64(amount6))
		return node.TransferContract(key, contract, data)
	}
	return "", fmt.Errorf("the type %s not support now", typs)
}

// 往外转 提币
func sendOut(contract, to string, amount decimal.Decimal) (string, error) {
	return send(mainAccout, contract, to, amount)
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
