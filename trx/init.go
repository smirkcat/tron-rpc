package trx

import (
	"context"
	"crypto/ecdsa"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tron/log"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
	"github.com/smirkcat/hdwallet"
)

var ctx, canceltask = context.WithCancel(context.Background())
var wg sync.WaitGroup

var port = "8245"
var trxdecimal int32 = 6

var IsMulti bool // 是否采用外部多链地址

var minScanBlock int64 = 58737696 // 最小 扫描高度
var targetHeight int64
var blockHeightTop int64
var minAmount decimal.Decimal
var remainAmount = decimal.New(10, 0) // 保留10个

var goroutineNumScan int64 = 4 // 扫描交易记录的并发携程数

var mainAddr = ""                // 主地址
var mainAccout *ecdsa.PrivateKey // 主地址密钥

// 归集参数
var minFee = decimal.New(5, 0)  // 每个地址至少保留多少trx手续费
var perFee = decimal.New(10, 0) // 每次归集每个合约需要手续费消耗

var dbengine *DB // 数据库连接

var globalConf GlobalConf

var curr = getCurrentDirectory() + `/`

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func InitSeed() {
	if globalConf.Seed != "" {
		hdwallet.Decrypt(globalConf.Seed, globalConf.SeedPri)
		hdwallet.InitHdwallet(globalConf.Seed)
	}
}

// InitLog 初始化日志文件
func InitLog() {
	var logConfigInfoName, logConfigErrorName, logLevel string
	logConfigInfoName = curr + "tron.log"
	logConfigErrorName = curr + "tron-err.log"
	logLevel = globalConf.LogLevel
	log.Init(logConfigInfoName, logConfigErrorName, logLevel)
}

func InitConfig() {
	if _, err := toml.DecodeFile(curr+"tron.toml", &globalConf); err != nil {
		panic(err)
	}
	if globalConf.Client.Port != "" {
		port = globalConf.Client.Port
	}

	IsMulti = globalConf.IsMulti

	minAmount = globalConf.Collection.MinAmount

	minScanBlock = globalConf.Scantraderecord.MinScanBlock

	if globalConf.Scantraderecord.GoroutineNum > 0 {
		goroutineNumScan = globalConf.Scantraderecord.GoroutineNum
	}

	if globalConf.Client.Count > 0 && globalConf.Client.Count < 100 {
		count = globalConf.Client.Count
	}
	// 最大100trx
	if globalConf.Client.Feelimit > 0 && globalConf.Client.Feelimit < 100000000 {
		feelimit = globalConf.Client.Feelimit
	}

	if globalConf.Client.MinFee.Cmp(decimal.Zero) > 0 {
		minFee = globalConf.Client.MinFee
	}

	if globalConf.Client.PerFee.Cmp(decimal.Zero) > 0 {
		perFee = globalConf.Client.PerFee
	}
}

func InitDB() {
	var err error
	dbengine, err = NewDB(globalConf.Client.DBAddr)
	if err != nil {
		panic(err)
	}
	err = dbengine.Sync()
	if err != nil {
		panic(err)
	}
	InitAddressDB(globalConf.Client.DBAddrMulti)
}

// InitMainAndFee 初始化主账户和手续费账户
func InitMainAndFee() {
	var err error
	mainAddr = globalConf.Client.MainAddr
	mainAccout, err = loadAccountWithUUID(globalConf.Client.MainPri, globalConf.Client.Password)
	if err != nil {
		panic(err)
	}
}

// InitWalletInfo 初始化钱包信息
func InitWalletInfo() {
	targetHeight = getlastBlock()
	log.Info("lastblock:", targetHeight)
	err := getWalletInfo()
	if err != nil {
		panic(err)
	}
	log.Info("walletInfo:", walletInfo)
}

// Init 初始化
func Init() {
	InitConfig()
	InitLog() // 首先初始化日志
	InitDB()
	InitMainNode(globalConf.Client.NodeTrx)
	InitAllNode(globalConf.Client.NodeUrl)
	InitContract(globalConf.Contracts)
	InitMainAndFee()
	InitWalletInfo()
	InitSeed()
	task()
}

func task() {
	var scanT = globalConf.Scantraderecord.TimeIntervalSec
	if scanT < 1 {
		scanT = 60
	}
	var collectT = globalConf.Collection.TimeIntervalMin
	if collectT < 1 {
		collectT = 30
	}
	var timed = time.Duration(scanT) * time.Second
	var timec = time.Duration(collectT) * time.Minute
	tiker := time.NewTicker(timed)
	tikerT := time.NewTicker(2 * timed)
	tikerC := time.NewTicker(timec)
	wg.Add(3)
	go func() {
		log.Info("RunTransaction Ticker")
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				tiker.Stop()
				return
			case <-tiker.C:
				log.Debug("start RunTransaction")
				RunTransaction()
				log.Debug("stop RunTransaction")
			}
		}
	}()
	go func() {
		log.Info("getWalletInfo Ticker")
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				tikerT.Stop()
				return
			case <-tikerT.C:
				log.Debug("start getWalletInfo")
				getWalletInfo()
				log.Debug("stop getWalletInfo")
			}
		}
	}()
	go func() {
		log.Info("RunCollect Ticker")
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				tikerC.Stop()
				return
			case <-tikerC.C:
				log.Debug("start RunCollect")
				RunCollect()
				log.Debug("stop RunCollect")
			}
		}
	}()
}
