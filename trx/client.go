package trx

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"tron/log"
	"tron/service"

	"github.com/smirkcat/hdwallet"

	wallet "tron/util"

	"github.com/shopspring/decimal"
)

var nodeall []string
var nodemain string

var mapContract = make(map[string]*Contract)
var mapContractType = map[string]bool{
	"trx":   true,
	"trc10": true,
	"trc20": true,
}

var walletInfo = wallet.Info{
	ContractBalance: make(map[string]json.Number),
}
var lockInfo sync.RWMutex

const (
	Trx   string = "trx"
	Trc10 string = "trc10"
	Trc20 string = "trc20"
)

// InitContract 初始化所有合约
func InitContract(contracts []Contract) {
	for i, v := range contracts {
		if ok := mapContractType[v.Type]; ok {
			mapContract[v.Contract] = &contracts[i]
		} else {
			panic(fmt.Errorf("the contract type %s is not exist pleasecheck", v.Type))
		}
	}
}

func InitMainNode(url string) {
	nodemain = url
	nodeall = make([]string, 1)
	nodeall[0] = url
}

func InitAllNode(url []string) {
	nodeall = append(nodeall, url...)
}

func newGrpcClient(url string) *service.GrpcClient {
	return service.NewGrpcClient(url)
}

// 后期改为 长链接
func getMaineNode() *service.GrpcClient {
	node := newGrpcClient(nodemain)
	err := node.Start()
	if err != nil {
		log.Errorf("main node %s err %s", nodemain, err.Error())
	}
	return node
}

// GenerateRandomNumber 生成count个[start,end)结束的不重复的随机数 为了用户抽奖 TODO
func GenerateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}
	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandOneNode() *service.GrpcClient {
	lens := len(nodeall)
	for i := 0; i < lens; i++ {
		nodeurl := nodeall[rand.Intn(lens)]
		node := newGrpcClient(nodeurl)
		err := node.Start()
		if err != nil {
			log.Errorf("node %s err %s", nodeurl, err.Error())
			continue
		}
		return node
	}
	log.Warn("use main node")
	return getMaineNode()
}

// 判断当前属于什么合约
func chargeContract(contract string) (string, int32) {
	if contract == "trx" || contract == "" {
		return Trx, trxdecimal
	}
	if v := mapContract[contract]; v != nil {
		if ok := mapContractType[v.Type]; ok {
			return v.Type, v.Decimal
		}
	}
	return "NONE", 18
}

// IsContract 判断当前合约是否存在
func IsContract(contract string) bool {
	if contract == "trx" || contract == "" {
		return true
	}
	if v := mapContract[contract]; v != nil {
		if ok := mapContractType[v.Type]; ok {
			return true
		}
	}
	return false
}

// GetWalletInfo 钱包信息
func getWalletInfoContract(contract string) wallet.Info {
	var info = walletInfo
	if contract != Trx {
		lockInfo.RLock()
		if v, ok := walletInfo.ContractBalance[contract]; ok {
			info.Balance = v
		} else {
			info.Balance = json.Number("0")
		}
		lockInfo.RUnlock()
	}
	return info
}

func processBlockHeight(block string) int64 {
	num := strings.Split(block, ",")[0]
	heights := strings.Split(num, ":")
	if len(heights) < 1 {
		return 0
	}
	height, _ := strconv.ParseInt(heights[1], 10, 64)
	return height
}

// GetWalletInfo 钱包信息
func getWalletInfo() (err error) {
	node := getMaineNode()
	defer node.Conn.Close()
	re, err1 := node.GetNodeInfo()
	if err1 != nil {
		err = err1
	} else {
		tmp := processBlockHeight(re.Block)
		if tmp > 0 {
			blockHeightTop = tmp
		} else {
			blockHeightTop = re.BeginSyncNum
		}
		walletInfo.BlockHeight = blockHeightTop
		walletInfo.Blocks = targetHeight
		walletInfo.Connections = int64(re.CurrentConnectCount)
		walletInfo.Difficulty = re.TotalFlow
	}
	ac, err1 := node.GetAccount(mainAddr)
	lockInfo.Lock()
	if err1 != nil {
		if err != nil {
			err = fmt.Errorf("node: %s balance: %s", err, err1)
		} else {
			err = err1
		}
	} else {
		walletInfo.Balance = json.Number(decimal.New(ac.Balance, -6).String())
		for _, v := range mapContract {
			if v.Type == Trc10 && ac.AssetV2 != nil {
				if vv, ok := ac.AssetV2[v.Contract]; ok {
					walletInfo.ContractBalance[v.Contract] = json.Number(decimal.New(vv, -v.Decimal).String())
					continue
				}
			}
		}
	}

	for _, v := range mapContract {
		if v.Type != Trc20 {
			continue
		}
		re, err := node.GetConstantResultOfContract(mainAccout, v.Contract, processBalanceOfParameter(mainAddr))
		if err != nil || len(re) < 1 {
			continue
		}
		walletInfo.ContractBalance[v.Contract] = json.Number(decimal.New(processBalanceOfData(re[0]), -v.Decimal).String())
	}
	lockInfo.Unlock()
	return
}

// 获取余额
func getBalanceByAddress(contract, addr string) (decimal.Decimal, error) {
	typs, decimalnum := chargeContract(contract)
	if typs == Trc20 {
		balance, err := getTrc20BalanceByAddress(contract, addr, mainAccout)
		return decimal.New(balance, -decimalnum), err
	}
	node := getMaineNode()
	defer node.Conn.Close()
	ac, err := node.GetAccount(addr)
	if err != nil {
		return decimal.Zero, err
	}
	switch typs {
	case Trx:
		return decimal.New(ac.Balance, -decimalnum), err
	case Trc10:
		if ac.AssetV2 != nil {
			if v, ok := ac.AssetV2[contract]; ok {
				return decimal.New(v, -decimalnum), err
			}
		}
	}
	return decimal.Zero, nil
}

func getTrc20BalanceByAddress(contract, addr string, ac *ecdsa.PrivateKey) (int64, error) {
	node := getMaineNode()
	defer node.Conn.Close()
	re, err := node.GetConstantResultOfContract(ac, contract, processBalanceOfParameter(addr))
	if err != nil || len(re) < 1 {
		return 0, err
	}
	return processBalanceOfData(re[0]), nil
}

func getlastBlock() int64 {
	if minScanBlock < 0 {
		return -minScanBlock
	}
	tmp := minScanBlock
	block, err := dbengine.LoadLastBlockHeight()
	if err != nil || block < tmp {
		block = tmp
	}
	return block
}

func validaddress(addr string) bool {
	if len(addr) != 34 {
		return false
	}
	if string(addr[0:1]) != "T" {
		return false
	}
	_, err := hdwallet.DecodeCheck(addr)
	return err == nil
}

func loadAccountWithUUID(pri, uuid string) (*ecdsa.PrivateKey, error) {
	pwd := hdwallet.HashAndSalt([]byte(uuid))
	return hdwallet.LoadPrivateKeyFromDecrypt(pri, pwd)
}

func loadAccount(addr string) (*ecdsa.PrivateKey, error) {
	re, err := SearchAccount(addr)
	if err != nil {
		return nil, err
	}
	if re == nil {
		return nil, fmt.Errorf("adderss %s is not exist", addr)
	}
	var pwd string
	if re.User != "" {
		pwd = hdwallet.HashAndSalt([]byte(re.User))
		return hdwallet.LoadPrivateKeyFromDecrypt(re.PrivateKey, pwd)
	}
	return hdwallet.GetPrivateKeyByHexString(re.PrivateKey)
}

func creataddress() (*Account, error) {
	var uuidv4 = hdwallet.GenPwd()
	pwd := hdwallet.HashAndSalt([]byte(uuidv4))
	index, privateKey, err := NewPrivateKey()
	if err != nil {
		return nil, err
	}
	adderss := hdwallet.PrikeyToAddressTron(privateKey)
	priEncrypt, err := hdwallet.StorePrivateKeyToDecrypt(privateKey, pwd)
	if err != nil {
		return nil, err
	}
	accountT := &Account{
		Address:    adderss,
		Index:      index,
		PrivateKey: priEncrypt,
		PublicKey:  hdwallet.PubkeyToHexString(privateKey.Public().(*ecdsa.PublicKey)),
		User:       uuidv4,
		Ctime:      time.Now().Unix(),
		Amount:     0,
	}
	_, err = dbengine.InsertAccount(accountT)
	return accountT, err
}
