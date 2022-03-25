package trx

import (
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

const (
	Send         = "send"          // 提币 即 主地址提币到其他地址
	Receive      = "receive"       // 本平台地址 分配的用户地址收到的
	ReceiveOther = "receive_other" // 主地址收到的外来地址转帐
	Collect      = "collect"       // 本平台地址归集到主地址
	CollectOwn   = "collect_own"   // 站内转账
	CollectSend  = "collect_send"  // 本平台地址提币到站外 异常的
)

// OtherParam .
type OtherParam struct {
	Key   string `xorm:"'key' unique"`
	Value string `xorm:"'value'"`
}

// TableName 表名
func (fh OtherParam) TableName() string {
	return "param"
}

// Account 账户分配的地址
type Account struct {
	ID         int64  `xorm:"'id' pk autoincr"`
	Address    string `xorm:"'address' unique DEFAULT '' "`         // 唯一索引
	PublicKey  string `xorm:"'public_key'"  json:"publickey"`       // 公钥新版字段 如果有就是新版
	PrivateKey string `xorm:"'private_key'" sql:"comment:'地址私钥'"`   // 地址私钥
	Index      int    `xorm:"'index' DEFAULT 0" sql:"comment:'位置'"` // 唯一
	User       string `xorm:"'user'"`
	Ctime      int64  `xorm:"'ctime'"`                          // 创建时间
	Amount     int64  `xorm:"'amount' index INTEGER DEFAULT 0"` // 主链币种余额
}

func (fh Account) TableName() string {
	return "account"
}

// Balance  代币余额
type Balance struct {
	ID       int64  `xorm:"'id' pk autoincr"`
	Address  string `xorm:"'address' DEFAULT '' "`
	Contract string `xorm:"'contract' index"` // 哪种合约
	Amount   int64  `xorm:"'amount' index INTEGER DEFAULT 0"`
}

func (fh Balance) TableName() string {
	return "balance"
}

// Transactions .
type Transactions struct {
	ID          int64  `xorm:"'id' pk autoincr" json:"-"`
	TxID        string `xorm:"'tx_id'" json:"txid"`
	BlockHeight int64  `xorm:"'block_height'" json:"blockheight"`
	PublicKey   string `xorm:"'public_key'"  json:"publickey"` // 公钥新版字段 如果有就是新版
	Address     string `xorm:"'address' index" json:""`
	FromAddress string `xorm:"'from_address'" json:"fromaddress"`
	Contract    string `xorm:"'contract' index"` // 哪种合约
	Amount      string `xorm:"'amount'" json:"amount"`
	Fee         string `xorm:"'fee'" json:"fee"` // 保留字段
	Timestamp   int64  `xorm:"'timestamp'"`
	Type        string `xorm:"'type'"` // send recive collect
}

// DB .
type DB struct {
	*xorm.Engine
}

// Close 关闭数据库引擎
func (db *DB) Close() {
	db.Engine.Close()
}

// Session 创建事务
func (db *DB) Session() *xorm.Session {
	return db.NewSession()
}

// NewDB 初始化数据库
func NewDB(url string) (*DB, error) {
	engine, err := xorm.NewEngine("sqlite3", url)
	return &DB{
		Engine: engine,
	}, err
}

// Sync 同步数据库结构
func (db *DB) Sync() error {
	return db.Sync2(new(Account), new(Transactions), new(OtherParam), new(Balance))
}

// InsertAccount 插入数据
func (db *DB) InsertAccount(account *Account) (int64, error) {
	return db.Cols("address", "private_key", "public_key", "index", "user", "ctime", "amount").Insert(account)
}

// UpdateAccount 更新数据
func (db *DB) UpdateAccount(account *Account) (int64, error) {
	return db.Where("address = ? ", account.Address).Cols("amount").Update(account)
}

// GetAccountWithAddr 搜索地址是否存在
func (db *DB) GetAccountWithAddr(addr string) (*Account, error) {
	var tmp Account
	ok, err := db.Where("address = ?", addr).Limit(1).Get(&tmp)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &tmp, err
}

func (db *DB) GetAccountMaxIndex() int {
	var resp map[string]int
	db.Table("account").Select("IFNULL(max(index),0) as maxid").Get(&resp)
	return resp["maxid"]
}

// GetAccount 获取所有账户
func (db *DB) GetAccount(from int) ([]Account, error) {
	var tmp = make([]Account, 0)
	if from < 0 {
		from = 0
	}
	err := db.Limit(1000, from).Find(&tmp)
	return tmp, err
}

// GetAccountWithBalance 获取大于minAmount的所有账户
func (db *DB) GetAccountWithBalance(startid int64, count int) ([]Account, error) {
	var tmp = make([]Account, 0)
	err := db.Where("id> ?", startid).Limit(count).Find(&tmp)
	return tmp, err
}

// SearchBalance 搜索余额记录是否存在
func (db *DB) SearchBalance(contract, address string) (*Balance, error) {
	var tmp Balance
	ok, err := db.Where("contract = ? and address =?", contract, address).Get(&tmp)
	if !ok || err != nil {
		return nil, err
	}
	return &tmp, err
}

// InsertBalance 插入数据
func (db *DB) InsertBalance(account *Balance) (int64, error) {
	re, _ := db.SearchBalance(account.Contract, account.Address)
	if re != nil {
		return db.Table(re).ID(re.ID).Update(map[string]interface{}{"amount": account.Amount})
	}
	return db.Insert(account)
}

// GetAccountWithContractBalance 获取大于minAmount的所有账户合约余额
func (db *DB) GetAccountWithContractBalance(contract string, minAmount, startid int64, count int) ([]Balance, error) {
	var tmp = make([]Balance, 0)
	err := db.Where("contract= ? and amount >= ? and id > ?", contract, minAmount, startid).Asc("id").Limit(count).Find(&tmp)
	return tmp, err
}

// GetSumContractBalance 获取合约总余额
func (db *DB) GetSumContractBalance(contract string) (map[string]int64, error) {
	var tmp = make(map[string]int64, 0)
	_, err := db.Table("balance").Select("sum(amount) as sumall").Where("contract= ?", contract).Get(&tmp)
	return tmp, err
}

// GetTransactions 获取最近交易记录
func (db *DB) GetTransactions(contract, addr string, count, skip int) ([]Transactions, error) {
	var tmp = make([]Transactions, 0)

	if count < 1 || count > 1000 {
		count = 300
	}
	if skip < 0 {
		skip = 0
	}
	tmpdb := db.Limit(count, skip).Where("(type=? OR type=?) And contract =? ", Send, Receive, contract)
	if addr != "*" && addr != "" {
		tmpdb = tmpdb.Where("addr = ?", addr)
	}
	err := tmpdb.Desc("id").Find(&tmp)
	return tmp, err
}

// GetCollestTransactions 获取指定时间段内归集交易记录
func (db *DB) GetCollestTransactions(sTime, eTime int64, contract string) ([]Transactions, error) {
	var tmp = make([]Transactions, 0)
	if eTime < sTime || eTime < 1 {
		eTime = 0
	}
	tmpdb := db.Where("type=? and contract=? and timesmap>=? ", Collect, contract, sTime)
	if eTime > 1 {
		tmpdb = tmpdb.Where("timesmap<=?", eTime)
	}
	err := tmpdb.Desc("id").Find(&tmp)
	return tmp, err
}

// SearchTransactions 搜索交易记录是否存在
func (db *DB) SearchTransactions(txid string) (*Transactions, error) {
	var tmp Transactions
	ok, err := db.Where("tx_id = ?", txid).Limit(1).Get(&tmp)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &tmp, err
}

// InsertTransactions 插入数据
func (db *DB) InsertTransactions(transactions *Transactions) (int64, error) {
	re, _ := db.SearchTransactions(transactions.TxID)
	if re != nil {
		return 0, nil
	}
	return db.Cols("tx_id", "block_height", "address", "public_key", "from_address", "contract",
		"amount", "fee", "timestamp", "type").Insert(transactions)
}

// LoadLastBlockHeight 获取最后一次扫描高度 已经扫描到这个高度
func (db *DB) LoadLastBlockHeight() (int64, error) {
	var tmp OtherParam
	ok, err := db.Where("key='block'").Limit(1).Get(&tmp)
	if err != nil || !ok {
		return 0, err
	}
	var un int64
	un, err = strconv.ParseInt(tmp.Value, 10, 0)
	return un, err
}

// InsertLastBlockHeight 更新最后一次扫描高度
func (db *DB) InsertLastBlockHeight(num int64) (err error) {
	var ok bool
	var tmp = OtherParam{
		Key: "block",
	}
	ok, err = db.Exist(&tmp)
	if err != nil || !ok {
		_, err = db.Insert(&tmp)
	} else {
		tmp.Value = strconv.FormatInt(num, 10)
		_, err = db.Where("key='block'").Cols("value").Update(&tmp)
	}
	return
}
