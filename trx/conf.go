package trx

import (
	"github.com/shopspring/decimal"
)

// GlobalConf .
type GlobalConf struct {
	Client          `toml:"client"`          // 钱包相关配置
	Scantraderecord `toml:"scantraderecord"` // 扫描交易记录配置
	Collection      `toml:"collection"`      // 钱包归集配置
	Contracts       []Contract               `toml:"contract"` // 合约
}

// Client 钱包节点
type Client struct {
	NodeTrx  string   `toml:"nodeTrx"`   // 超级节点url 主节点
	NodeUrl  []string `toml:"nodeUrl"`   // 其他节点url
	Password string   `toml:"password"`  // 主钱包密码 加密后的
	MainAddr string   `toml:"main_addr"` // 主钱包地址
	DBAddr   string   `toml:"db_addr"`   // sqlite 地址
	KeyStore string   `toml:"key_store"` // 用户钱包保存路径
	Port     string   `toml:"port"`      // 监听端口 trx
	LogLevel string   `toml:"logLevel"`  // 日志等级
}

// Scantraderecord 扫描交易记录配置
type Scantraderecord struct {
	TimeIntervalSec int64 `toml:"time_interval_sec"` // 扫描间隔
	GoroutineNum    int64 `toml:"goroutine_num"`     // 扫描协程数
	MinScanBlock    int64 `toml:"min_scan_block"`    // 扫描交易记录起始位置 如果配置为正数 取最大值开始扫描 如果为负数 则取绝对值 从绝对值位置开始扫描，不取最大值开始扫描
}

// Collection 归集配置
type Collection struct {
	TimeIntervalMin int64           `toml:"time_interval_min"` // 所有代币归集扫描间隔
	MinAmount       decimal.Decimal `toml:"min_amount"`        // trx 最小归集数目
	//GoroutineNum    int             `toml:"goroutine_num"`
}

// Contract 合约 TRC20 和 TRC10
type Contract struct {
	Port                string          `toml:"port"`       // rpc 监听端口
	Name                string          `toml:"name"`       // USDT BTT
	Type                string          `toml:"type"`       // TRC20 和 TRC10
	Contract            string          `toml:"contract"`   // 合约地址或者合约ID
	Issuer              string          `toml:"issuer"`     // 合约创建者
	CollectionMinAmount decimal.Decimal `toml:"min_amount"` // 代币最小归集数目
}
