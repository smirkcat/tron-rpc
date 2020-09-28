package trx

import (
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
	"tron/common/base58"
	"tron/common/crypto"
	"tron/log"
	"tron/service"

	wallet "tron/util"

	"github.com/gofrs/uuid"
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
func InitContract(contracts []Contract) error {
	for i, v := range contracts {
		if ok, _ := mapContractType[v.Type]; ok {
			mapContract[v.Contract] = &contracts[i]
		} else {
			return fmt.Errorf("the contract type %s is not exist pleasecheck", v.Type)
		}
	}
	return nil
}

func InitMainNode(url string) error {
	nodemain = url
	nodeall = make([]string, 1)
	nodeall[0] = url
	return nil
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

// InitLog 初始化日志文件
func InitLog() {
	var logConfigInfoName, logConfigErrorName, logLevel string
	logConfigInfoName = curr + "trx.log"
	logConfigErrorName = curr + "trx-err.log"
	logLevel = globalConf.LogLevel
	log.Init(logConfigInfoName, logConfigErrorName, logLevel)
}

// 判断当前属于什么合约
func chargeContract(contract string) (string, int32) {
	if contract == "trx" || contract == "" {
		return Trx, trxdecimal
	}
	if v := mapContract[contract]; v != nil {
		if ok, _ := mapContractType[v.Type]; ok {
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
		if ok, _ := mapContractType[v.Type]; ok {
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

// GetWalletInfo 钱包信息
func getWalletInfo() (err error) {
	node := getMaineNode()
	defer node.Conn.Close()
	re, err1 := node.GetNodeInfo()
	if err1 != nil {
		err = err1
	} else {
		blockHeightTop = re.BeginSyncNum
		walletInfo.BlockHeight = blockHeightTop
		walletInfo.Blocks = targetHeight
		walletInfo.Connections = int64(re.CurrentConnectCount)
		walletInfo.Difficulty = re.TotalFlow
	}
	ac, err1 := node.GetAccount(mainAddr)
	lockInfo.Lock()
	if err1 != nil {
		if err != nil {
			err = fmt.Errorf("node: %s\nbalance: %s\n", err, err1)
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

func getNodeInfo() error {
	node := getMaineNode()
	defer node.Conn.Close()
	re, err := node.GetNodeInfo()
	if err != nil {
		return err
	}
	blockHeightTop = re.BeginSyncNum
	walletInfo.BlockHeight = blockHeightTop
	walletInfo.Blocks = blockHeightTop
	walletInfo.Connections = int64(re.CurrentConnectCount)
	walletInfo.Difficulty = re.TotalFlow
	return nil
}

// trx
func getBalance() error {
	node := getMaineNode()
	defer node.Conn.Close()
	ac, err := node.GetAccount(mainAddr)
	if err != nil {
		return err
	}
	walletInfo.Balance = json.Number(decimal.New(ac.Balance, -6).String())
	return nil
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

func getBlockHeight() error {
	node := getMaineNode()
	defer node.Conn.Close()
	block, err := node.GetNowBlock()
	if err != nil {
		return err
	}
	processBlock(block)
	blockHeightTop = block.BlockHeader.RawData.Number
	walletInfo.BlockHeight = blockHeightTop
	walletInfo.Blocks = blockHeightTop
	return nil
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
	_, err := base58.DecodeCheck(addr)
	if err != nil {
		return false
	}
	return true
}

func loadAccountWithUUID(addr, uuid string) (*ecdsa.PrivateKey, error) {
	pwd := hashAndSalt([]byte(uuid + "trx"))
	return loadAccountFile(keystore+"/"+addr, pwd)
}

func loadAccount(addr string) (*ecdsa.PrivateKey, error) {
	re, err := dbengine.SearchAccount(addr)
	if err != nil {
		return nil, err
	}
	if re == nil {
		return nil, fmt.Errorf("adderss %s is not exist", addr)
	}
	pwd := hashAndSalt([]byte(re.User + "trx"))
	return loadAccountFile(keystore+"/"+addr, pwd)
}

func loadAccountFile(filePath, pwd string) (account *ecdsa.PrivateKey, err error) {
	b, err1 := ioutil.ReadFile(filePath)
	if err1 != nil {
		err = err1
		return
	}
	re, err1 := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		err = err1
		return
	}
	md5sum := md5.Sum([]byte(pwd))
	result, err1 := AesDecrypt(re, md5sum[:])
	if err != nil {
		err = err1
		return
	}
	account, err = crypto.GetPrivateKeyByHexString(string(result))
	return
}

// storeAccountToKeyStoreFile store an account to a json file
func storeAccountToKeyStoreFile(account *ecdsa.PrivateKey, password, walletName string) (filePath string, err error) {
	filePath = walletName
	f, err1 := os.Create(filePath) //创建文件
	if err1 != nil {
		err = err1
		return
	}
	defer f.Close()
	prikey := crypto.PrikeyToHexString(account)
	md5sum := md5.Sum([]byte(password))
	result, err1 := AesEncrypt([]byte(prikey), md5sum[:])
	if err1 != nil {
		err = err1
		return
	}
	_, err = f.WriteString(base64.StdEncoding.EncodeToString(result))
	return
}

func creataddress() (string, error) {
	var uuidv4 = uuid.Must(uuid.NewV4()).String()
	p := hashAndSalt([]byte(uuidv4 + "trx"))
	re, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}
	addr := base58.EncodeCheck(crypto.PubkeyToAddress(re.PublicKey).Bytes())
	_, err = storeAccountToKeyStoreFile(re, p, keystore+"/"+addr)
	if err != nil {
		return "", err
	}
	accountT := &Account{
		Address: addr,
		User:    uuidv4,
		Ctime:   time.Now().Unix(),
		Amount:  0,
	}
	_, err = dbengine.InsertAccount(accountT)
	return addr, err
}

func hashAndSalt(pwd []byte) string {
	lens := len(pwd)
	var num int
	for i := 0; i < lens; i++ {
		tmp := int(pwd[i])
		if tmp != 45 {
			num += tmp
		}
	}
	num = num%15 + 1
	var k = num
	var pass = make([]byte, 16)

	for i := 0; i < 16; i++ {
		pass[i] = te4[k]
		k += num
	}
	return hex.EncodeToString(pass) + "trx"
}

// 来源 c++密码源代码
var te4 = [256]byte{
	0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
	0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
	0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
	0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
	0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
	0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
	0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
	0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
	0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
	0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
	0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
	0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
	0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
	0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
	0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
	0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16}
