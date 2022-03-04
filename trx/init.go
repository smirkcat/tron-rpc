package trx

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tron/log"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
)

var ctx, canceltask = context.WithCancel(context.Background())
var wg sync.WaitGroup

var port = "8245"
var trxdecimal int32 = 6

var IsMulti bool // 是否采用外部多链地址

var minScanBlock int64 = 23513066 // 最小 扫描高度
var targetHeight int64
var blockHeightTop int64
var minAmount decimal.Decimal
var remainAmount = decimal.New(10, 0) // 保留10个

var goroutineNumScan int64 = 4 // 扫描交易记录的并发携程数

var mainAddr = ""                // 主地址
var mainAccout *ecdsa.PrivateKey // 主地址密钥

// 归集参数
var minFee = decimal.New(3, 0) // 每个地址至少保留多少trx手续费
var perFee = decimal.New(3, 0) // 每次归集每个合约需要手续费消耗

// 为了替换前一个
var mainAddr1 = ""                // 主地址2
var mainAccout1 *ecdsa.PrivateKey // 主地址密钥2
var istwomain bool                // 是否两个主地址 第一个主地址还是负责提币和转手续费 第二个的地址负责收集归集币

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

func InitConfig() {
	if _, err := toml.DecodeFile(curr+"tron.toml", &globalConf); err != nil {
		fmt.Println(err)
		_, err = toml.Decode(string(getConfig()), &globalConf)
		if err != nil {
			panic(err)
		}
	}
	if globalConf.Client.Port != "" {
		port = globalConf.Client.Port
	}

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
	// mainAddr1 = globalConf.Client.MainAddr1
	// if mainAddr1 != "" {
	// 	mainAccout1, err = loadAccountWithUUID(mainAddr1, globalConf.Client.Password1)
	// 	if err != nil {
	// 		log.Error(err)
	// 	}
	// 	istwomain = true
	// }
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

//获取默认的数据库配置
func getConfig() []byte {
	return []byte(`
# grpc.trongrid.io:50051 正式
# grpc.shasta.trongrid.io:50051 测试
[client]
nodeTrx="grpc.trongrid.io:50051"
main_addr="TQCknYutmcMxGoq32JqQWvn1MzyRfuQirC" #主钱包地址
password="eb1804aa-fa7d-4782-8145-afe4da83c56d" #主钱包秘钥加密前的密码 uuid
main_pri="" #主钱包地址加密私钥
db_addr="D:/go/tron/trx/tron.db"
port="8245"
logLevel="info" # 日志等级默认
count=3 #批量查询交易记录个数
feelimit=5000000 # 每次转账trc20合约燃烧的能量 单位sun 默认5trx
perfee=5 # 每次归集每个合约需要手续费消耗
minfee=5 # 每个地址至少保留多少trx手续费

# 合约配置 
[[contract]]
name="USDT"  # 暂时没有用到
type="trc20" # 合约类型
contract="TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" # trc20 合约地址
issuer="THPvaUhoh2Qn2y9THCZML3H815hhFhn5YC" # 发行者地址 暂时没有用到
port="8246" # 监听端口
min_amount=5 # 最小归集数量
decimal=6 # 币种小数位

# [[contract]]
# name="BTT"
# type="trc10" # 合约类型
# contract="1002000" # 合约配置 trc10 合约ID assertname
# issuer="TF5Bn4cJCT6GVeUgyCN4rBhDg42KBrpAjg" # 发行者地址 暂时没有用到
# port="8247" # 监听端口
# min_amount=0.1 # 最小归集数量

[collection]
time_interval_min=60 # 归集检测间隔  单位 分
min_amount=10 # 最小归集钱包余额 单位TRX 后面 6个零 1TRX =10^6 SUN

[scantraderecord]
time_interval_sec=5 # 扫描交易记录检测间隔 单位秒
# 扫描交易记录起始位置 如果配置为正数 
# 如果为负数 则取绝对值 从绝对值位置开始扫描，不取最大值开始扫描
min_scan_block = 23520251
goroutine_num=4 # 每次扫描开的协程数量
`)
}
