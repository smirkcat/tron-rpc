package trx

import (
	"sync"
	"time"
	"tron/log"

	"github.com/shopspring/decimal"
)

// 每次批量查询最多100个 100个会body很大 所以这里限制20个
var count int64 = 20

// RunTransaction .
func RunTransaction() {
	tmpendheight := blockHeightTop
	dvalue := tmpendheight - targetHeight
	if dvalue < 1 {
		return
	}
	countnum := dvalue / count // 总共单位请求数量
	if dvalue%count > 0 {
		countnum++
	}
	daverage := countnum / goroutineNumScan // 循环次数

	remaning := countnum % goroutineNumScan // 最后一次的次数

	blockHeight := targetHeight

	var wg = &sync.WaitGroup{}
	var i int64 = 0
	for ; i < daverage; i++ {
		var j int64 = 0
		for ; j < goroutineNumScan; j++ {
			wg.Add(1)
			go func(start, end int64) {
				getBlockWithHeights(start, end)
				wg.Done()
			}(blockHeight, blockHeight+count)
			blockHeight += count
		}
		wg.Wait()
		if blockHeight > tmpendheight {
			targetHeight = tmpendheight
		} else {
			targetHeight = blockHeight
		}
		err := dbengine.InsertLastBlockHeight(targetHeight)
		log.Infof("insert height %d err %v", targetHeight, err)
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
	if remaning > 0 {
		for i = 0; i < remaning; i++ {
			wg.Add(1)
			go func(start, end int64) {
				getBlockWithHeights(start, end)
				wg.Done()
			}(blockHeight, blockHeight+count)
			blockHeight += count
		}
		wg.Wait()
		targetHeight = tmpendheight
		err := dbengine.InsertLastBlockHeight(targetHeight)
		log.Infof("insert height %d err %v", targetHeight, err)
	}
}

func collect(contract, addr string) {
	// 成交归集检测
	_, err := loadAccount(addr)

	if err != nil {
		log.Errorf("loadAccount contract %s addr %s err: %s", contract, addr, err.Error())
		return
	}

	amount, err := getBalanceByAddress(contract, addr)
	if err != nil {
		log.Errorf("getBalance contract %s addr %s err: %s", contract, addr, err.Error())
		return
	}
	minamount := minAmount

	typs, decimalnum := chargeContract(contract)

	if typs != Trx {
		minamount = mapContract[contract].CollectionMinAmount
	}

	log.Infof("contract %s, addr %s amount %s getBalance before", contract, addr, amount)

	if amount.GreaterThanOrEqual(minamount) {
		amounttrx, err := getBalanceByAddress("", addr)
		if err != nil {
			log.Errorf("getBalance trx %s addr %s err: %s", addr, err.Error())
			return
		}

		if amounttrx.LessThan(decimal.NewFromFloat(0.5)) {
			txid, err := sendFee(addr, decimal.NewFromFloat(1))
			if err != nil {
				log.Errorf("send fee %d addr %s err: %s", 1, addr, err.Error())
			} else {
				log.Infof("send fee %d addr %s txid: %s", 1, addr, txid)
				// 波场一般3秒一个块 12个确认的话 就是 36秒 等待36秒
				time.Sleep(36 * time.Second)
				// 这里不做进一步判断余额了
			}
		}
		txid, err := sendIn(contract, addr, amount)
		if err != nil {
			log.Errorf("collect contract %s addr %s err: %s", contract, addr, err.Error())
		} else {
			log.Infof("contract %s addr %s the collect txid: %s", contract, addr, txid)
		}
		amountt, err := getBalanceByAddress(contract, addr)
		if err != nil {
			log.Errorf("getBalance contract %s addr %s err: %s", contract, addr, err.Error())
		} else {
			amount = amountt
			log.Infof("getBalance contract %s addr %s is %s", contract, addr, amountt.String())
		}
	}

	var amountdecimal = decimal.New(1, decimalnum)
	amountac, _ := amount.Mul(amountdecimal).Float64()

	if contract == Trx {
		var tmp = &Account{
			Address: addr,
			Amount:  int64(amountac),
		}
		_, err = dbengine.UpdateAccount(tmp)
		if err != nil {
			log.Errorf("UpdateAccount %v err: %s", *tmp, err)
		}
	} else {
		var tmp = &Balance{
			Address:  addr,
			Contract: contract,
			Amount:   int64(amountac),
		}
		_, err = dbengine.InsertBalance(tmp)
		if err != nil {
			log.Errorf("UpdateBalance %v err: %s", *tmp, err)
		}
	}
}

// RunCollect 获取数据库中 大于指定余额钱包余额进行归集
func RunCollect() {
	// 合约归集检测
	for _, v := range mapContract {
		amountdecimal := decimal.New(1, v.Decimal)
		amountContract6, _ := v.CollectionMinAmount.Mul(amountdecimal).Float64()
		addr2, err := dbengine.GetAccountWithContractBalance(v.Contract, int64(amountContract6), 1000)
		if err != nil {
			log.Errorf("GetAccountWithContractBalance err:%s", err)
		} else {
			var lens = len(addr2)
			for i := 0; i < lens; i++ {
				collect(v.Contract, addr2[i].Address)
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}

	}
	// num60 := decimal.New(1, 6)
	// amount6, _ := minAmount.Mul(num60).Float64()
	// // trx 归集检测 每次检测1000个满足条件的 然后等待下次检测
	// addr, err := dbengine.GetAccountWithBalance(int64(amount6), 1000)
	// if err != nil {
	// 	log.Errorf("GetAccountWithBalance err:%s", err)
	// } else {
	// 	var lens = len(addr)
	// 	for i := 0; i < lens; i++ {
	// 		collect(Trx, addr[i].Address)
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 		}
	// 	}
	// }
}
