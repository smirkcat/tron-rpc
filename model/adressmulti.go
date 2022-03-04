package model

import (
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

var dbengine *DB // 数据库连接

func InitDB(url string) {
	var err error
	dbengine, err = NewDB(url)
	if err != nil {
		panic(err)
	}
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
	engine, err := xorm.NewEngine("mysql", url)
	return &DB{
		Engine: engine,
	}, err
}

// 新版分配地址池保存表

const TN_ADDRESS_multi = `bgd_address_multi`

type AddressMulti struct {
	Id          int64  `xorm:"'id' pk autoincr"`
	PrivateKey  string `xorm:"'private_key'" sql:"comment:'地址私钥'"` // 唯一性 加密
	PublicKey   string `xorm:"'public_key'" sql:"comment:'地址公钥'"`
	Account     string `xorm:"'account'" sql:"comment:'地址标签'"` // 可能加秘钥或者其他注释
	EthAddress  string `xorm:"'eth_address'" sql:"comment:'ETH系列地址'"`
	TronAddress string `xorm:"'tron_address'" sql:"comment:'Tron系列地址'"`
	Ext         string `xorm:"'ext'" sql:"comment:'拓展字段'"`
	Index       int    `xorm:"'index'" sql:"comment:'位置'"` // 唯一
	Status      uint8  `xorm:"'status'" sql:"comment:'是否已经推送'"`
	CreateTime  int64  `xorm:"'create_time'" sql:"comment:'创建时间'"`
}

func (AddressMulti) TableName() string {
	return TN_ADDRESS_multi
}

// GetUsedOfAddress 查询某个币种地址存在否
func GetUsedOfAddressMulti(coinName, address string) (AddressMulti, error) {
	var resp AddressMulti
	err := dbengine.Table(TN_ADDRESS_multi).Select("*").
		Where("? = ?", coinName+"_address", address).Find(&resp)
	return resp, err
}

// GetUsedOfPrivateKey 查询某个私钥存在
func GetUsedOfPrivateKey(pri string) AddressMulti {
	var resp AddressMulti
	dbengine.Table(TN_ADDRESS_multi).Select("*").
		Where("private_key = ", pri).Find(&resp)
	return resp
}

// GetUsedOfPrivateKey 查询某个公钥是否存在
func GetUsedOfPublickey(pub string) AddressMulti {
	var resp AddressMulti
	dbengine.Table(TN_ADDRESS_multi).Select("*").
		Where("public_key = ", pub).Find(&resp)
	return resp
}
