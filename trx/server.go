package trx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tron/util"

	"github.com/semrush/zenrpc/v2"
	"github.com/shopspring/decimal"
)

// Service Trx 钱包信息
type Service struct {
	zenrpc.Service
	Contract string
	Port     string
}

// Getinfo  获取钱包信息
func (as Service) Getinfo() util.Info {
	return getWalletInfoContract(as.Contract)
}

// GetNewAddress  获取新地址
func (as Service) GetNewAddress() (string, error) {
	ac, err := creataddress()
	if err != nil {
		return "", err
	}
	return ac.Address, nil
}

// ValidateAddress 校验地址
func (as Service) ValidateAddress(addr string) util.ValidateAddress {
	var resp util.ValidateAddress
	resp.IsValidate = validaddress(addr)
	return resp
}

// ListTransactions 获取指定地址最近的交易记录
//zenrpc:count=300
//zenrpc:skip=0
//zenrpc:addr="*"
func (as Service) ListTransactions(addr string, count, skip int) ([]util.Transactions, error) {
	return recentTransactions(as.Contract, addr, count, skip)
}

// SendToAddress 提币请求
func (as Service) SendToAddress(addr string, amount json.Number) (string, error) {
	amountt, _ := decimal.NewFromString(string(amount))
	return sendOut(as.Contract, addr, amountt)
}

// GetRecords  归集交易记录 中转记录
func (as Service) GetRecords(sTime, eTime int64) ([]util.SummaryData, error) {
	return collectTransactions(as.Contract, sTime, eTime)
}

//go:generate zenrpc

// Serv 监听服务
func Serv(ctx context.Context, rpcSev Service) *http.Server {
	server := &http.Server{}
	rpc := zenrpc.NewServer(zenrpc.Options{ExposeSMD: true})
	rpc.Register("", rpcSev)
	//rpc.Use(zenrpc.Logger(log.New(os.Stderr, "", log.LstdFlags)))
	httpw := http.NewServeMux()
	httpw.Handle("/", rpc)
	server.Handler = httpw
	server.Addr = ":" + rpcSev.Port
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()
	return server
}

// InitAllContarctServer 初始化所有节点
func InitAllContarctServer(ctx context.Context, exit chan<- struct{}) {
	var servers []*http.Server
	var service = Service{
		Contract: Trx,
		Port:     port,
	}
	servers = append(servers, Serv(ctx, service))
	for _, v := range mapContract {
		service.Contract = v.Contract
		service.Port = v.Port
		servers = append(servers, Serv(ctx, service))
	}
	go func() {
		<-ctx.Done()
		for _, v := range servers {
			v.Close()
		}
		canceltask()
		wg.Wait()
		exit <- struct{}{}
	}()
}
