package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"time"
	"tron/api"
	"tron/core"
	"tron/hexutil"
	"tron/util"

	"github.com/smirkcat/hdwallet"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	Address string
	Conn    *grpc.ClientConn
	Client  api.WalletClient
}

func NewGrpcClient(address string) *GrpcClient {
	client := new(GrpcClient)
	client.Address = address
	return client
}

func (g *GrpcClient) Start() error {
	var err error
	g.Conn, err = grpc.Dial(g.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	g.Client = api.NewWalletClient(g.Conn)
	return nil
}

// 新版
func ContextTimeout(sec int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*time.Duration(sec))
}

func (g *GrpcClient) ListWitnesses() (*api.WitnessList, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	witnessList, err := g.Client.ListWitnesses(ctx,
		new(api.EmptyMessage))

	if err != nil {
		return nil, fmt.Errorf("get witnesses error: %v", err)
	}

	return witnessList, nil
}

func (g *GrpcClient) ListNodes() (*api.NodeList, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	nodeList, err := g.Client.ListNodes(ctx,
		new(api.EmptyMessage))
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %v", err)
	}
	return nodeList, nil
}

func (g *GrpcClient) GetNodeInfo() (*core.NodeInfo, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	node, err := g.Client.GetNodeInfo(ctx, new(api.EmptyMessage))
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (g *GrpcClient) GetAccount(address string) (*core.Account, error) {
	account := new(core.Account)
	var err error
	account.Address, err = hdwallet.DecodeCheck(address)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetAccount(ctx, account)
	return result, err
}

func (g *GrpcClient) GetNowBlock() (*api.BlockExtention, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetNowBlock2(ctx, new(api.EmptyMessage))
	return result, err
}

func (g *GrpcClient) GetAssetIssueByAccount(address string) (*api.AssetIssueList, error) {
	account := new(core.Account)
	account.Address, _ = hdwallet.DecodeCheck(address)
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	result, err := g.Client.GetAssetIssueByAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("get asset issue by account error: %v", err)
	}
	return result, nil
}

func (g *GrpcClient) GetNextMaintenanceTime() (*api.NumberMessage, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetNextMaintenanceTime(ctx, new(api.EmptyMessage))
	if err != nil {
		return nil, fmt.Errorf("get next maintenance time error: %v", err)
	}
	return result, nil
}

func (g *GrpcClient) TotalTransaction() (*api.NumberMessage, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	return g.Client.TotalTransaction(ctx, new(api.EmptyMessage))
}

func (g *GrpcClient) GetAccountNet(address string) (*api.AccountNetMessage, error) {
	account := new(core.Account)

	account.Address, _ = hdwallet.DecodeCheck(address)
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	return g.Client.GetAccountNet(ctx, account)
}

func (g *GrpcClient) GetAssetIssueByName(name string) (*core.AssetIssueContract, error) {

	assetName := new(api.BytesMessage)
	assetName.Value = []byte(name)
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	return g.Client.GetAssetIssueByName(ctx, assetName)
}

func (g *GrpcClient) GetBlockByNum(num int64) (*api.BlockExtention, error) {
	numMessage := new(api.NumberMessage)
	numMessage.Num = num
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	result, err := g.Client.GetBlockByNum2(ctx, numMessage)
	return result, err
}

func (g *GrpcClient) GetBlockById(id string) (*core.Block, error) {
	blockId := new(api.BytesMessage)
	var err error

	blockId.Value, err = hexutil.Decode(id)

	if err != nil {
		return nil, fmt.Errorf("get block by id error: %v", err)
	}
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetBlockById(ctx, blockId)

	if err != nil {
		return nil, fmt.Errorf("get block by id error: %v", err)
	}
	return result, nil
}

func (g *GrpcClient) GetAssetIssueList() (*api.AssetIssueList, error) {
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	return g.Client.GetAssetIssueList(ctx, new(api.EmptyMessage))
}

func (g *GrpcClient) GetBlockByLimitNext(start, end int64) (*api.BlockListExtention, error) {
	blockLimit := new(api.BlockLimit)
	blockLimit.StartNum = start
	blockLimit.EndNum = end
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	return g.Client.GetBlockByLimitNext2(ctx, blockLimit)
}

func (g *GrpcClient) GetTransactionById(id string) (*core.Transaction, error) {
	transactionId := new(api.BytesMessage)
	var err error
	transactionId.Value, err = hexutil.Decode(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	return g.Client.GetTransactionById(ctx, transactionId)
}

func (g *GrpcClient) GetTransactionInfoById(id string) (*core.TransactionInfo, error) {
	transactionId := new(api.BytesMessage)
	var err error
	transactionId.Value, err = hexutil.Decode(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetTransactionInfoById(ctx, transactionId)
	return result, err
}

func (g *GrpcClient) GetBlockByLatestNum(num int64) (*api.BlockListExtention, error) {
	numMessage := new(api.NumberMessage)
	numMessage.Num = num
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	result, err := g.Client.GetBlockByLatestNum2(ctx, numMessage)
	return result, err
}

func (g *GrpcClient) CreateAccount(ownerKey *ecdsa.PrivateKey,
	accountAddress string) (*api.Return, error) {

	accountCreateContract := new(core.AccountCreateContract)
	accountCreateContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.PublicKey).Bytes()
	accountCreateContract.AccountAddress, _ = hdwallet.DecodeCheck(accountAddress)

	ctx, cancel := ContextTimeout(30)
	defer cancel()
	accountCreateTransaction, err := g.Client.CreateAccount(ctx, accountCreateContract)

	if err != nil {
		return nil, fmt.Errorf("create account error: %v", err)
	}
	if accountCreateTransaction == nil ||
		len(accountCreateTransaction.GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("create account error: invalid transaction")
	}

	util.SignTransaction(accountCreateTransaction, ownerKey)
	return g.Client.BroadcastTransaction(ctx,
		accountCreateTransaction)
}

func (g *GrpcClient) UpdateAccount(ownerKey *ecdsa.PrivateKey,
	accountName string) (*api.Return, error) {
	var err error
	accountUpdateContract := new(core.AccountUpdateContract)
	accountUpdateContract.AccountName = []byte(accountName)
	accountUpdateContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	accountUpdateTransaction, err := g.Client.UpdateAccount(ctx, accountUpdateContract)
	if err != nil {
		return nil, fmt.Errorf("update account error: %v", err)
	}
	if accountUpdateTransaction == nil ||
		len(accountUpdateTransaction.GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("update account error: invalid transaction")
	}

	util.SignTransaction(accountUpdateTransaction, ownerKey)

	return g.Client.BroadcastTransaction(ctx, accountUpdateTransaction)
}

func (g *GrpcClient) Transfer(ownerKey *ecdsa.PrivateKey, toAddress string, amount int64) (string, error) {
	transferContract := new(core.TransferContract)
	transferContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ToAddress, _ = hdwallet.DecodeCheck(toAddress)
	transferContract.Amount = amount
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	transferTransactionEx, err := g.Client.CreateTransaction2(ctx, transferContract)

	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil ||
		len(transferTransaction.GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	hash, err := util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(hash)

	result, err := g.Client.BroadcastTransaction(ctx,
		transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %s", result.String())
	}
	return txid, err
}

func (g *GrpcClient) TransferAsset(ownerKey *ecdsa.PrivateKey, AssetName, toAddress string, amount int64) (string, error) {
	transferContract := new(core.TransferAssetContract)
	transferContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ToAddress, _ = hdwallet.DecodeCheck(toAddress)
	transferContract.AssetName, _ = hdwallet.DecodeCheck(AssetName)
	transferContract.Amount = amount
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	transferTransactionEx, err := g.Client.TransferAsset2(ctx, transferContract)

	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil ||
		len(transferTransaction.GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	hash, err := util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(hash)

	result, err := g.Client.BroadcastTransaction(ctx, transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %s", result.String())
	}
	return txid, err
}

func (g *GrpcClient) TransferContract(ownerKey *ecdsa.PrivateKey, Contract string, data []byte, feeLimit int64) (string, error) {
	transferContract := new(core.TriggerSmartContract)
	transferContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ContractAddress, _ = hdwallet.DecodeCheck(Contract)
	transferContract.Data = data
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	transferTransactionEx, err := g.Client.TriggerConstantContract(ctx, transferContract)
	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil ||
		len(transferTransaction.GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	if feeLimit > 0 {
		transferTransaction.RawData.FeeLimit = feeLimit
	}

	hash, err := util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(hash)

	result, err := g.Client.BroadcastTransaction(ctx,
		transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %s", result.String())
	}
	return txid, err
}

func (g *GrpcClient) GetConstantResultOfContract(ownerKey *ecdsa.PrivateKey, Contract string, data []byte) ([][]byte, error) {
	transferContract := new(core.TriggerSmartContract)
	transferContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.PublicKey).Bytes()
	transferContract.ContractAddress, _ = hdwallet.DecodeCheck(Contract)
	transferContract.Data = data
	ctx, cancel := ContextTimeout(20)
	defer cancel()
	transferTransactionEx, err := g.Client.TriggerConstantContract(ctx, transferContract)
	if err != nil {
		return nil, err
	}
	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		return nil, fmt.Errorf("GetConstantResult error: invalid TriggerConstantContract")
	}
	return transferTransactionEx.GetConstantResult(), err
}

func (g *GrpcClient) FreezeBalance(ownerKey *ecdsa.PrivateKey,
	frozenBalance, frozenDuration int64) (*api.Return, error) {
	freezeBalanceContract := new(core.FreezeBalanceContract)
	freezeBalanceContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()
	freezeBalanceContract.FrozenBalance = frozenBalance
	freezeBalanceContract.FrozenDuration = frozenDuration
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	freezeBalanceTransaction, err := g.Client.FreezeBalance(ctx, freezeBalanceContract)

	if err != nil {
		return nil, fmt.Errorf("freeze balance error: %v", err)
	}

	if freezeBalanceTransaction == nil || len(freezeBalanceTransaction.
		GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("freeze balance error: invalid transaction")
	}

	util.SignTransaction(freezeBalanceTransaction, ownerKey)

	return g.Client.BroadcastTransaction(ctx, freezeBalanceTransaction)
}

func (g *GrpcClient) UnfreezeBalance(ownerKey *ecdsa.PrivateKey) (*api.Return, error) {
	unfreezeBalanceContract := new(core.UnfreezeBalanceContract)
	unfreezeBalanceContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.PublicKey).Bytes()
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	unfreezeBalanceTransaction, err := g.Client.UnfreezeBalance(ctx, unfreezeBalanceContract)

	if err != nil {
		return nil, fmt.Errorf("unfreeze balance error: %v", err)
	}

	if unfreezeBalanceTransaction == nil ||
		len(unfreezeBalanceTransaction.GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("unfreeze balance error: invalid transaction")
	}

	util.SignTransaction(unfreezeBalanceTransaction, ownerKey)
	return g.Client.BroadcastTransaction(ctx, unfreezeBalanceTransaction)
}

func (g *GrpcClient) CreateAssetIssue(ownerKey *ecdsa.PrivateKey,
	name, description, urlStr string, totalSupply, startTime, endTime,
	FreeAssetNetLimit,
	PublicFreeAssetNetLimit int64, trxNum,
	icoNum, voteScore int32, frozenSupply map[string]string) (*api.Return, error) {
	assetIssueContract := new(core.AssetIssueContract)

	assetIssueContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()

	assetIssueContract.Name = []byte(name)

	if totalSupply <= 0 {
		return nil, fmt.Errorf("create asset issue error: total supply <= 0")
	}
	assetIssueContract.TotalSupply = totalSupply

	if trxNum <= 0 {
		return nil, fmt.Errorf("create asset issue error: trxNum <= 0")
	}
	assetIssueContract.TrxNum = trxNum

	if icoNum <= 0 {
		return nil, fmt.Errorf("create asset issue error: num <= 0")
	}
	assetIssueContract.Num = icoNum

	now := time.Now().UnixNano() / 1000000
	if startTime <= now {
		return nil, fmt.Errorf("create asset issue error: start time <= current time")
	}
	assetIssueContract.StartTime = startTime

	if endTime <= startTime {
		return nil, fmt.Errorf("create asset issue error: end time <= start time")
	}
	assetIssueContract.EndTime = endTime

	if FreeAssetNetLimit < 0 {
		return nil, fmt.Errorf("create asset issue error: free asset net limit < 0")
	}
	assetIssueContract.FreeAssetNetLimit = FreeAssetNetLimit

	if PublicFreeAssetNetLimit < 0 {
		return nil, fmt.Errorf("create asset issue error: public free asset net limit < 0")
	}
	assetIssueContract.PublicFreeAssetNetLimit = PublicFreeAssetNetLimit

	assetIssueContract.VoteScore = voteScore
	assetIssueContract.Description = []byte(description)
	assetIssueContract.Url = []byte(urlStr)

	for key, value := range frozenSupply {
		amount, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("create asset issue error: convert error: %v", err)
		}
		days, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("create asset issue error: convert error: %v", err)
		}
		assetIssueContractFrozenSupply := new(core.
			AssetIssueContract_FrozenSupply)
		assetIssueContractFrozenSupply.FrozenAmount = amount
		assetIssueContractFrozenSupply.FrozenDays = days
		assetIssueContract.FrozenSupply = append(assetIssueContract.
			FrozenSupply, assetIssueContractFrozenSupply)
	}
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	assetIssueTransaction, err := g.Client.CreateAssetIssue(ctx, assetIssueContract)

	if err != nil {
		return nil, fmt.Errorf("create asset issue error: %v", err)
	}

	if assetIssueTransaction == nil || len(assetIssueTransaction.
		GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("create asset issue error: invalid transaction")
	}

	util.SignTransaction(assetIssueTransaction, ownerKey)

	return g.Client.BroadcastTransaction(ctx, assetIssueTransaction)
}

func (g *GrpcClient) UpdateAssetIssue(ownerKey *ecdsa.PrivateKey,
	description, urlStr string,
	newLimit, newPublicLimit int64) (*api.Return, error) {

	updateAssetContract := new(core.UpdateAssetContract)

	updateAssetContract.OwnerAddress = hdwallet.PubkeyToTronAddress(ownerKey.
		PublicKey).Bytes()

	updateAssetContract.Description = []byte(description)
	updateAssetContract.Url = []byte(urlStr)
	updateAssetContract.NewLimit = newLimit
	updateAssetContract.NewPublicLimit = newPublicLimit
	ctx, cancel := ContextTimeout(30)
	defer cancel()
	updateAssetTransaction, err := g.Client.UpdateAsset(ctx, updateAssetContract)

	if err != nil {
		return nil, fmt.Errorf("update asset issue error: %v", err)
	}

	if updateAssetTransaction == nil || len(updateAssetTransaction.
		GetRawData().GetContract()) == 0 {
		return nil, fmt.Errorf("update asset issue error: invalid transaction")
	}

	util.SignTransaction(updateAssetTransaction, ownerKey)

	return g.Client.BroadcastTransaction(ctx,
		updateAssetTransaction)
}
