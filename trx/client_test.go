package trx

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/smirkcat/hdwallet"
)

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
	t.Log(validaddress("TDRPyn57F4riYTJFcHaQbrzgFaGe8HSumL"))
}

func TestNowHeight(t *testing.T) {
	Init()
	//getBlockHeight()
	//getBlockWithHeight(991186)
	t.Log(getBlockWithHeight(15003338))
	//getBlockWithHeight(991186)
	//t.Log(getBlockWithHeights(1075923, 1076023))
}

func TestGetWalletInfo(t *testing.T) {
	Init()
	//t.Log(getNodeInfo())
	t.Log(getBalanceByAddress("", "TAUN6FwrnwwmaEqYcckffC7wYmbaS6cBiX"))
}

func TestMain(t *testing.T) {
	Init()
	select {}
}

func TestTrans(t *testing.T) {
	Init()
	t.Log("fdhjhjd")
	txid, err := sendIn("trx", "TLVYpQH98E9hQXDTgAbnCLwTWquaupxz3T", decimal.NewFromFloat(0.01001))
	if err != nil {
		t.Log(err)
	}
	t.Log(txid)
}

func TestContractBalanceTrc20(t *testing.T) {
	Init()
	t.Log(getTrc20BalanceByAddress("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "TLXgWgXnJZVjL46Qyju52n6QVJRUQBoZyU", mainAccout))
}

func TestNode(t *testing.T) {
	// 14.104.81.238:54004
	node := newGrpcClient("14.104.83.38:54004")
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
