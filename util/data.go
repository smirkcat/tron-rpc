package util

import (
	"encoding/json"
)

// Info TRX 钱包信息
type Info struct {
	Version         int
	ProtocolVersion int
	WalletVersion   int
	Balance         json.Number
	Difficulty      int64
	BlockHeight     int64
	Blocks          int64
	Connections     int64
	TimeOffset      int64
	Time            int64
	ContractBalance map[string]json.Number `json:"-"`
}

// Transactions ..
type Transactions struct {
	Account       string      `json:"account"`
	TxID          string      `json:"txid"`
	Address       string      `json:"address"`
	PublicKey     string      `json:"publickey"` // 公钥新版字段 如果有就是新版
	FromAddress   string      `json:"fromaddress"`
	Category      string      `json:"category"`
	Amount        json.Number `json:"amount"`
	Fee           json.Number `json:"fee"`
	Vout          int         `json:"vout"`
	Confirmations int64       `json:"confirmations"`
	Generated     bool        `json:"generated"`
	BlockHash     string      `json:"blockhash"`
	BlockIndex    int64       `json:"blockindex"`
	BlockTime     int64       `json:"blocktime"`
	Time          int64       `json:"time"`
	TimeReceived  int64       `json:"timereceived"`
}

// SummaryData 归集中转记录
type SummaryData struct {
	TxID         string `json:"txid"`
	Account      string `json:"account"`
	Address      string `json:"address"`
	PublicKey    string `json:"publickey"` // 公钥新版字段 如果有就是新版
	FromAddress  string `json:"fromaddress"`
	Amount       string `json:"amount"`
	BlockIndex   int64  `json:"blockindex"`
	Blocktime    int64  `json:"blocktime"`
	Category     string `json:"category"`
	Fee          string `json:"fee"`
	Time         int64  `json:"time"`
	TimeReceived int64  `json:"timeReceived"`
}

// ValidateAddress 钱包合法检测
type ValidateAddress struct {
	IsValidate bool `json:"isvalid"`
}

// Address 钱包地址
type Address struct {
	StandardAddress string `json:"standard_address"`
	PaymentID       string `json:"payment_id"`
}

//IntegrateAddress string `json:"integrated_address"`
