package trx

import (
	"fmt"
	"math/big"
	"testing"
	"tron/common/base58"
	"tron/common/hexutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
)

func TestCreatAccount(t *testing.T) {
	Init()
	addr, err := creataddress()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(addr)
	re, err := loadAccount(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(re)
	// pwd := hashAndSalt([]byte("1070fcd0-ed20-425d-af1f-6d217d2e4820trx"))
	// md5sum := md5.Sum([]byte(pwd))
	// result, err1 := AesEncrypt([]byte("d2e5300b3a4069f7c525b957f3668c1672914fa3e5f73ada83773ed5b6616a0a"), md5sum[:])
	// fmt.Println(base64.StdEncoding.EncodeToString(result), err1)
	// cont, err1 := AesDecrypt(result, md5sum[:])
	// fmt.Println(string(cont), err1)
}

func TestValid(t *testing.T) {
	fmt.Println(validaddress("TDRPyn57F4riYTJFcHaQbrzgFaGe8HSumL"))
}

func TestNowHeight(t *testing.T) {
	Init()
	//getBlockHeight()
	//getBlockWithHeight(991186)
	fmt.Println(getBlockWithHeight(15003338))
	//getBlockWithHeight(991186)
	//fmt.Println(getBlockWithHeights(1075923, 1076023))
}

func TestGetWalletInfo(t *testing.T) {
	Init()
	//fmt.Println(getNodeInfo())
	fmt.Println(getBalanceByAddress("", "TAUN6FwrnwwmaEqYcckffC7wYmbaS6cBiX"))
}

func TestMain(t *testing.T) {
	Init()
	select {}
}

func TestTrans(t *testing.T) {
	Init()
	fmt.Println("fdhjhjd")
	txid, err := sendIn("trx", "TLVYpQH98E9hQXDTgAbnCLwTWquaupxz3T", decimal.NewFromFloat(0.01001))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(txid)
}

func TestContractBalanceTrc20(t *testing.T) {
	Init()
	fmt.Println(getTrc20BalanceByAddress("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "TLXgWgXnJZVjL46Qyju52n6QVJRUQBoZyU", mainAccout))
}

func TestTokenAddress(t *testing.T) {
	transferFnSignature := []byte("balanceOf(address)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID))

	addr, err := base58.DecodeCheck("TLXgWgXnJZVjL46Qyju52n6QVJRUQBoZyU")
	if err != nil {
		fmt.Println(err)
	}
	paddedAddress := common.LeftPadBytes(addr, 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 00000000000000000000004173d5888eedd05efeda5bca710982d9c13b975f98

	amount, _ := new(big.Int).SetString("10000000", 10)

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount))

	// "eRAT Token","ERAT",2000000000,8
	var data []byte
	a1 := common.LeftPadBytes([]byte("eRAT Token"), 32)
	a2 := common.LeftPadBytes([]byte("ERAT"), 32)
	amount3 := new(big.Int).SetInt64(2000000000)
	a3 := common.LeftPadBytes(amount3.Bytes(), 32)
	amount4 := new(big.Int).SetInt64(8)
	a4 := common.LeftPadBytes(amount4.Bytes(), 32)
	data = append(data, a1...)
	data = append(data, a2...)
	data = append(data, a3...)
	data = append(data, a4...)
	fmt.Println(hexutil.Encode(data))
}

func TestNode(t *testing.T) {
	// 14.104.81.238:54004
	node := newGrpcClient("14.104.83.38:54004")
	err := node.Start()
	if err != nil {
		fmt.Println(err)
	}
	defer node.Conn.Close()
	re, err := node.GetNodeInfo()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(re)
}
