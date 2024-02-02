package trx

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/smirkcat/hdwallet"
)

func init() {
	curr = "./"
}
func TestCreatAddress(t *testing.T) {
	var uuidv4 = hdwallet.GenPwd()
	pwd := hdwallet.HashAndSalt([]byte(uuidv4))
	privateKey, err := hdwallet.NewPrivateKeyIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	adderss := hdwallet.PrikeyToAddressTron(privateKey)
	priEncrypt, err := hdwallet.StorePrivateKeyToDecrypt(privateKey, pwd)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(uuidv4)
	t.Log(adderss)
	t.Log(priEncrypt)
}

var dbaddr = "tron.db"

func TestDB(t *testing.T) {
	var err error
	dbengine, err = NewDB(dbaddr)
	if err != nil {
		panic(err)
	}
	err = dbengine.Sync()
	if err != nil {
		panic(err)
	}
}

func TestCreatAccount(t *testing.T) {
	Init()
	addr, err := creataddress()
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(addr)
	re, err := loadAccount(addr.Address)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(re)
}

func TestValid(t *testing.T) {
	t.Log(validaddress("TL4kyKaXJ9gThBhHtyMSN4ZMKSaD5cZUGL"))
}

func TestNowHeight(t *testing.T) {
	Init()
	t.Log(getBlockWithHeight(43578000))
}

func TestGetWalletInfo(t *testing.T) {
	Init()
	//t.Log(getNodeInfo())
	t.Log(getBalanceByAddress("", "TL4kyKaXJ9gThBhHtyMSN4ZMKSaD5cZUGL"))
}

func TestMain(t *testing.T) {
	Init()
	select {}
}

func TestTrans(t *testing.T) {
	Init()
	txid, err := sendIn("trx", "TL4kyKaXJ9gThBhHtyMSN4ZMKSaD5cZUGL", decimal.NewFromFloat(0.01001))
	if err != nil {
		t.Log(err)
	}
	t.Log(txid)
}

func TestContractBalanceTrc20(t *testing.T) {
	Init()
	t.Log(getTrc20BalanceByAddress("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "TL4kyKaXJ9gThBhHtyMSN4ZMKSaD5cZUGL", mainAccout))
}

func TestNode(t *testing.T) {
	node := newGrpcClient("grpc.trongrid.io:50051")
	err := node.Start()
	if err != nil {
		t.Log(err)
	}
	defer node.Conn.Close()
	re, err := node.GetNodeInfo()
	if err != nil {
		t.Log(err)
	}
	t.Log(re)
}
