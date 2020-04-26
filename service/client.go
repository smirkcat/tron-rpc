package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"time"
	"tron/api"
	"tron/common/base58"
	"tron/common/crypto"
	"tron/common/hexutil"
	"tron/core"
	"tron/util"
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
	g.Conn, err = grpc.Dial(g.Address, grpc.WithInsecure())
	if err != nil {
		return err
	}
	g.Client = api.NewWalletClient(g.Conn)
	return nil
}

func timeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*15)
	return ctx
}

func (g *GrpcClient) ListWitnesses() *api.WitnessList {
	witnessList, err := g.Client.ListWitnesses(timeoutContext(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("get witnesses error: %v\n", err)
	}

	return witnessList
}

func (g *GrpcClient) ListNodes() *api.NodeList {
	nodeList, err := g.Client.ListNodes(timeoutContext(),
		new(api.EmptyMessage))
	if err != nil {
		log.Fatalf("get nodes error: %v\n", err)
	}
	return nodeList
}

func (g *GrpcClient) GetNodeInfo() (*core.NodeInfo, error) {
	node, err := g.Client.GetNodeInfo(timeoutContext(), new(api.EmptyMessage))
	if err != nil {
		return nil, err
	}
	return node, err
}

func (g *GrpcClient) GetAccount(address string) (*core.Account, error) {
	account := new(core.Account)
	var err error
	account.Address, err = base58.DecodeCheck(address)
	if err != nil {
		return nil, err
	}
	result, err := g.Client.GetAccount(timeoutContext(), account)
	return result, err
}

func (g *GrpcClient) GetNowBlock() (*api.BlockExtention, error) {
	result, err := g.Client.GetNowBlock2(timeoutContext(), new(api.EmptyMessage))
	return result, err
}

func (g *GrpcClient) GetAssetIssueByAccount(address string) *api.AssetIssueList {
	account := new(core.Account)

	account.Address, _ = base58.DecodeCheck(address)

	result, err := g.Client.GetAssetIssueByAccount(timeoutContext(),
		account)

	if err != nil {
		log.Fatalf("get asset issue by account error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetNextMaintenanceTime() *api.NumberMessage {

	result, err := g.Client.GetNextMaintenanceTime(timeoutContext(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("get next maintenance time error: %v", err)
	}

	return result
}

func (g *GrpcClient) TotalTransaction() *api.NumberMessage {

	result, err := g.Client.TotalTransaction(timeoutContext(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("total transaction error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetAccountNet(address string) *api.AccountNetMessage {
	account := new(core.Account)

	account.Address, _ = base58.DecodeCheck(address)

	result, err := g.Client.GetAccountNet(timeoutContext(), account)

	if err != nil {
		log.Fatalf("get account net error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetAssetIssueByName(name string) *core.AssetIssueContract {

	assetName := new(api.BytesMessage)
	assetName.Value = []byte(name)

	result, err := g.Client.GetAssetIssueByName(timeoutContext(), assetName)

	if err != nil {
		log.Fatalf("get asset issue by name error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetBlockByNum(num int64) (*api.BlockExtention, error) {
	numMessage := new(api.NumberMessage)
	numMessage.Num = num
	result, err := g.Client.GetBlockByNum2(timeoutContext(), numMessage)
	return result, err
}

func (g *GrpcClient) GetBlockById(id string) *core.Block {
	blockId := new(api.BytesMessage)
	var err error

	blockId.Value, err = hexutil.Decode(id)

	if err != nil {
		log.Fatalf("get block by id error: %v", err)
	}

	result, err := g.Client.GetBlockById(timeoutContext(), blockId)

	if err != nil {
		log.Fatalf("get block by id error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetAssetIssueList() *api.AssetIssueList {

	result, err := g.Client.GetAssetIssueList(timeoutContext(), new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("get asset issue list error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetBlockByLimitNext(start, end int64) (*api.BlockListExtention, error) {
	blockLimit := new(api.BlockLimit)
	blockLimit.StartNum = start
	blockLimit.EndNum = end
	result, err := g.Client.GetBlockByLimitNext2(timeoutContext(), blockLimit)
	return result, err
}

func (g *GrpcClient) GetTransactionById(id string) (*core.Transaction, error) {
	transactionId := new(api.BytesMessage)
	var err error
	transactionId.Value, err = hexutil.Decode(id)
	if err != nil {
		return nil, err
	}
	result, err := g.Client.GetTransactionById(timeoutContext(), transactionId)
	return result, err
}

func (g *GrpcClient) GetTransactionInfoById(id string) (*core.TransactionInfo, error) {
	transactionId := new(api.BytesMessage)
	var err error
	transactionId.Value, err = hexutil.Decode(id)
	if err != nil {
		return nil, err
	}
	result, err := g.Client.GetTransactionInfoById(timeoutContext(), transactionId)
	return result, err
}

func (g *GrpcClient) GetBlockByLatestNum(num int64) (*api.BlockListExtention, error) {
	numMessage := new(api.NumberMessage)
	numMessage.Num = num
	result, err := g.Client.GetBlockByLatestNum2(timeoutContext(), numMessage)
	return result, err
}

func (g *GrpcClient) CreateAccount(ownerKey *ecdsa.PrivateKey,
	accountAddress string) *api.Return {

	accountCreateContract := new(core.AccountCreateContract)
	accountCreateContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.PublicKey).Bytes()
	accountCreateContract.AccountAddress, _ = base58.DecodeCheck(accountAddress)

	accountCreateTransaction, err := g.Client.CreateAccount(timeoutContext(), accountCreateContract)

	if err != nil {
		log.Fatalf("create account error: %v", err)
	}

	if accountCreateTransaction == nil || len(accountCreateTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("create account error: invalid transaction")
	}

	util.SignTransaction(accountCreateTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		accountCreateTransaction)

	if err != nil {
		log.Fatalf("create account error: %v", err)
	}

	return result
}

func (g *GrpcClient) UpdateAccount(ownerKey *ecdsa.PrivateKey,
	accountName string) *api.Return {
	var err error
	accountUpdateContract := new(core.AccountUpdateContract)
	accountUpdateContract.AccountName = []byte(accountName)
	accountUpdateContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()

	accountUpdateTransaction, err := g.Client.UpdateAccount(timeoutContext(), accountUpdateContract)

	if err != nil {
		log.Fatalf("update account error: %v", err)
	}

	if accountUpdateTransaction == nil || len(accountUpdateTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("update account error: invalid transaction")
	}

	util.SignTransaction(accountUpdateTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		accountUpdateTransaction)

	if err != nil {
		log.Fatalf("update account error: %v", err)
	}

	return result
}

func (g *GrpcClient) Transfer(ownerKey *ecdsa.PrivateKey, toAddress string, amount int64) (string, error) {
	transferContract := new(core.TransferContract)
	transferContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ToAddress, _ = base58.DecodeCheck(toAddress)
	transferContract.Amount = amount

	transferTransactionEx, err := g.Client.CreateTransaction2(timeoutContext(), transferContract)

	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil || len(transferTransaction.
		GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	err = util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(transferTransactionEx.Txid)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %v", result.String())
	}
	return txid, err
}

func (g *GrpcClient) TransferAsset(ownerKey *ecdsa.PrivateKey, AssetName, toAddress string, amount int64) (string, error) {
	transferContract := new(core.TransferAssetContract)
	transferContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ToAddress, _ = base58.DecodeCheck(toAddress)
	transferContract.AssetName, _ = base58.DecodeCheck(AssetName)
	transferContract.Amount = amount

	transferTransactionEx, err := g.Client.TransferAsset2(timeoutContext(), transferContract)

	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil || len(transferTransaction.
		GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	err = util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(transferTransactionEx.Txid)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %v", result.String())
	}
	return txid, err
}

func (g *GrpcClient) TransferContract(ownerKey *ecdsa.PrivateKey, Contract string, data []byte) (string, error) {
	transferContract := new(core.TriggerSmartContract)
	transferContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()
	transferContract.ContractAddress, _ = base58.DecodeCheck(Contract)
	transferContract.Data = data
	transferTransactionEx, err := g.Client.TriggerConstantContract(timeoutContext(), transferContract)
	var txid string
	if err != nil {
		return txid, err
	}
	transferTransaction := transferTransactionEx.Transaction
	if transferTransaction == nil || len(transferTransaction.
		GetRawData().GetContract()) == 0 {
		return txid, fmt.Errorf("transfer error: invalid transaction")
	}
	err = util.SignTransaction(transferTransaction, ownerKey)
	if err != nil {
		return txid, err
	}
	txid = hexutil.Encode(transferTransactionEx.Txid)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		transferTransaction)
	if err != nil {
		return "", err
	}
	if !result.Result {
		return "", fmt.Errorf("api get false the msg: %v", result.String())
	}
	return txid, err
}

func (g *GrpcClient) GetConstantResultOfContract(ownerKey *ecdsa.PrivateKey, Contract string, data []byte) ([][]byte, error) {
	transferContract := new(core.TriggerSmartContract)
	transferContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.PublicKey).Bytes()
	transferContract.ContractAddress, _ = base58.DecodeCheck(Contract)
	transferContract.Data = data
	transferTransactionEx, err := g.Client.TriggerConstantContract(timeoutContext(), transferContract)
	if err != nil {
		return nil, err
	}
	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		return nil, fmt.Errorf("GetConstantResult error: invalid TriggerConstantContract")
	}
	return transferTransactionEx.GetConstantResult(), err
}

func (g *GrpcClient) FreezeBalance(ownerKey *ecdsa.PrivateKey,
	frozenBalance, frozenDuration int64) *api.Return {
	freezeBalanceContract := new(core.FreezeBalanceContract)
	freezeBalanceContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()
	freezeBalanceContract.FrozenBalance = frozenBalance
	freezeBalanceContract.FrozenDuration = frozenDuration

	freezeBalanceTransaction, err := g.Client.FreezeBalance(timeoutContext(), freezeBalanceContract)

	if err != nil {
		log.Fatalf("freeze balance error: %v", err)
	}

	if freezeBalanceTransaction == nil || len(freezeBalanceTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("freeze balance error: invalid transaction")
	}

	util.SignTransaction(freezeBalanceTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		freezeBalanceTransaction)

	if err != nil {
		log.Fatalf("freeze balance error: %v", err)
	}

	return result
}

func (g *GrpcClient) UnfreezeBalance(ownerKey *ecdsa.PrivateKey) *api.Return {
	unfreezeBalanceContract := new(core.UnfreezeBalanceContract)
	unfreezeBalanceContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.PublicKey).Bytes()

	unfreezeBalanceTransaction, err := g.Client.UnfreezeBalance(timeoutContext(), unfreezeBalanceContract)

	if err != nil {
		log.Fatalf("unfreeze balance error: %v", err)
	}

	if unfreezeBalanceTransaction == nil || len(unfreezeBalanceTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("unfreeze balance error: invalid transaction")
	}

	util.SignTransaction(unfreezeBalanceTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		unfreezeBalanceTransaction)

	if err != nil {
		log.Fatalf("unfreeze balance error: %v", err)
	}

	return result
}

func (g *GrpcClient) CreateAssetIssue(ownerKey *ecdsa.PrivateKey,
	name, description, urlStr string, totalSupply, startTime, endTime,
	FreeAssetNetLimit,
	PublicFreeAssetNetLimit int64, trxNum,
	icoNum, voteScore int32, frozenSupply map[string]string) *api.Return {
	assetIssueContract := new(core.AssetIssueContract)

	assetIssueContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()

	assetIssueContract.Name = []byte(name)

	if totalSupply <= 0 {
		log.Fatalf("create asset issue error: total supply <= 0")
	}
	assetIssueContract.TotalSupply = totalSupply

	if trxNum <= 0 {
		log.Fatalf("create asset issue error: trxNum <= 0")
	}
	assetIssueContract.TrxNum = trxNum

	if icoNum <= 0 {
		log.Fatalf("create asset issue error: num <= 0")
	}
	assetIssueContract.Num = icoNum

	now := time.Now().UnixNano() / 1000000
	if startTime <= now {
		log.Fatalf("create asset issue error: start time <= current time")
	}
	assetIssueContract.StartTime = startTime

	if endTime <= startTime {
		log.Fatalf("create asset issue error: end time <= start time")
	}
	assetIssueContract.EndTime = endTime

	if FreeAssetNetLimit < 0 {
		log.Fatalf("create asset issue error: free asset net limit < 0")
	}
	assetIssueContract.FreeAssetNetLimit = FreeAssetNetLimit

	if PublicFreeAssetNetLimit < 0 {
		log.Fatalf("create asset issue error: public free asset net limit < 0")
	}
	assetIssueContract.PublicFreeAssetNetLimit = PublicFreeAssetNetLimit

	assetIssueContract.VoteScore = voteScore
	assetIssueContract.Description = []byte(description)
	assetIssueContract.Url = []byte(urlStr)

	for key, value := range frozenSupply {
		amount, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Fatalf("create asset issue error: convert error: %v", err)
		}
		days, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			log.Fatalf("create asset issue error: convert error: %v", err)
		}
		assetIssueContractFrozenSupply := new(core.
			AssetIssueContract_FrozenSupply)
		assetIssueContractFrozenSupply.FrozenAmount = amount
		assetIssueContractFrozenSupply.FrozenDays = days
		assetIssueContract.FrozenSupply = append(assetIssueContract.
			FrozenSupply, assetIssueContractFrozenSupply)
	}

	assetIssueTransaction, err := g.Client.CreateAssetIssue(timeoutContext(), assetIssueContract)

	if err != nil {
		log.Fatalf("create asset issue error: %v", err)
	}

	if assetIssueTransaction == nil || len(assetIssueTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("create asset issue error: invalid transaction")
	}

	util.SignTransaction(assetIssueTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		assetIssueTransaction)

	if err != nil {
		log.Fatalf("create asset issue error: %v", err)
	}

	return result
}

func (g *GrpcClient) UpdateAssetIssue(ownerKey *ecdsa.PrivateKey,
	description, urlStr string,
	newLimit, newPublicLimit int64) *api.Return {

	updateAssetContract := new(core.UpdateAssetContract)

	updateAssetContract.OwnerAddress = crypto.PubkeyToAddress(ownerKey.
		PublicKey).Bytes()

	updateAssetContract.Description = []byte(description)
	updateAssetContract.Url = []byte(urlStr)
	updateAssetContract.NewLimit = newLimit
	updateAssetContract.NewPublicLimit = newPublicLimit

	updateAssetTransaction, err := g.Client.UpdateAsset(timeoutContext(), updateAssetContract)

	if err != nil {
		log.Fatalf("update asset issue error: %v", err)
	}

	if updateAssetTransaction == nil || len(updateAssetTransaction.
		GetRawData().GetContract()) == 0 {
		log.Fatalf("update asset issue error: invalid transaction")
	}

	util.SignTransaction(updateAssetTransaction, ownerKey)

	result, err := g.Client.BroadcastTransaction(timeoutContext(),
		updateAssetTransaction)

	if err != nil {
		log.Fatalf("update asset issue error: %v", err)
	}

	return result
}
