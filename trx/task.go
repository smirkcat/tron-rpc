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

// 成交归集检测
func collectall(addr string) {
	// 成交归集检测
	_, err := loadAccount(addr)

	if err != nil {
		log.Errorf("loadAccount  addr %s err: %s", addr, err.Error())
		return
	}

	// 获取波场余额
	amounttrx, err := getBalanceByAddress("", addr)
	if err != nil {
		log.Errorf("getBalance trx %s addr %s err: %s", addr, err.Error())
		return
	}
	// 检测满足归集的合约余额数量标记
	var coll = make(map[string]decimal.Decimal, 0)
	for _, v := range mapContract {
		amount, err := getBalanceByAddress(v.Contract, addr)
		if err != nil {
			log.Errorf("getBalance contract %s addr %s err: %s", v.Contract, addr, err.Error())
			continue
		}
		log.Infof("getBalance contract %s addr %s is %s", v.Contract, addr, amount.String())
		if amount.GreaterThanOrEqual(v.CollectionMinAmount) {
			coll[v.Contract] = amount
		}
		// 更新余额
		var amountdecimal = decimal.New(1, v.Decimal)
		amountac, _ := amount.Mul(amountdecimal).Float64()
		var tmp = &Balance{
			Address:  addr,
			Contract: v.Contract,
			Amount:   int64(amountac),
		}
		_, err = dbengine.InsertBalance(tmp)
		if err != nil {
			log.Errorf("UpdateBalance %v err: %s", *tmp, err)
		}
	}
	// 有归集的合约
	if len(coll) > 0 {
		// 合计每个合约预留trx数量
		feeAmount := decimal.NewFromInt32(int32(len(coll))).Mul(perFee)

		if amounttrx.LessThan(feeAmount) {
			transfee := feeAmount.Add(minFee).Sub(amounttrx)
			txid, err := sendFee(addr, transfee)
			if err != nil {
				log.Errorf("send fee %s addr %s err: %s", transfee.String(), addr, err.Error())
			} else {
				log.Infof("send fee %s addr %s txid: %s", transfee.String(), addr, txid)
				// 波场一般3秒一个块 12个确认的话 就是 36秒 等待36秒
				time.Sleep(36 * time.Second)
				// 这里不做进一步判断余额了
			}
		}
		for contract, v := range coll {
			txid, err := sendIn(contract, addr, v)
			if err != nil {
				// 有可能是成功了的
				log.Errorf("contract %s addr %s the collect txid: %s err: %s", contract, addr, txid, err.Error())
			} else {
				log.Infof("contract %s addr %s the collect txid: %s", contract, addr, txid)
			}
			time.Sleep(10 * time.Second)
			amount, err := getBalanceByAddress(contract, addr)
			if err != nil {
				log.Errorf("getBalance contract %s addr %s err: %s", contract, addr, err.Error())
				continue
			}
			log.Infof("getBalance contract %s addr %s is %s", contract, addr, amount.String())
			// 更新余额
			var amountdecimal = decimal.New(1, mapContract[contract].Decimal)
			amountac, _ := amount.Mul(amountdecimal).Float64()
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

	// 再次获取波场余额 是否归集
	amounttrx, err = getBalanceByAddress("", addr)
	if err != nil {
		log.Errorf("getBalance trx %s addr %s err: %s", addr, err.Error())
		return
	}
	// 满足最小归集量 满足预留最小数量
	if amounttrx.GreaterThan(minAmount) && amounttrx.GreaterThan(remainAmount) {
		v := amounttrx.Sub(remainAmount)
		txid, err := sendIn("", addr, v)
		if err != nil {
			// 有可能是成功了的
			log.Errorf("collect trx %s addr %s err: %s", addr, err.Error())
		} else {
			log.Infof("trx  addr %s the collect txid: %s", addr, txid)
		}
		time.Sleep(10 * time.Second)
		amounttrx, err = getBalanceByAddress("", addr)
		if err != nil {
			log.Errorf("getBalance trx %s addr %s err: %s", addr, err.Error())
			return
		}
	}
	// 更新余额
	amountac, _ := amounttrx.Mul(decimal.New(1, 6)).Float64()
	var tmp = &Balance{
		Address:  addr,
		Contract: "",
		Amount:   int64(amountac),
	}
	_, err = dbengine.InsertBalance(tmp)
	if err != nil {
		log.Errorf("UpdateBalance %v err: %s", *tmp, err)
	}
}

// RunCollect 获取数据库中 大于指定余额钱包余额进行归集
func RunCollect() {
	// 归集检测 并行100个任务
	var task = make(chan bool, 100)
	var wgcollect sync.WaitGroup // 保持等待所有任务结束
	defer wgcollect.Wait()
	count := 1000
	var minAmountv int64

	for _, v := range mapContract {
		var startid int64
		tmp, _ := v.CollectionMinAmount.Mul(decimal.New(1, v.Decimal)).Float64()
		minAmountv = int64(tmp)
		for {
			addr, err := dbengine.GetAccountWithContractBalance(v.Contract, minAmountv, startid, count)
			if err != nil {
				log.Errorf("GetAccountWithContractBalance err:%v", err)
				select {
				case <-ctx.Done():
					return
				default:
				}
				break
			}
			var lens = len(addr)
			log.Infof("collect Contract %s nums %d", v.Contract, lens)
			for i := 0; i < lens; i++ {
				startid = addr[i].ID
				wgcollect.Add(1)
				task <- true
				go func(k int) {
					collectall(addr[k].Address)
					wgcollect.Done()
					<-task
				}(i)
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			if lens < count {
				break
			}
			time.Sleep(time.Second)
		}
		time.Sleep(time.Second)
	}

	var id int64 = 0
	tmp, _ := minAmount.Mul(decimal.New(1, 6)).Float64()
	minAmountv = int64(tmp)
	for {
		addr, err := dbengine.GetAccountWithContractBalance("", minAmountv, id, count)
		if err != nil {
			log.Errorf("GetAccountWithContractBalance err:%v", err)
			select {
			case <-ctx.Done():
				return
			default:
			}
			break
		}
		var lens = len(addr)
		log.Infof("collect trx nums %d", lens)
		if lens < 1 {
			return
		}

		for i := 0; i < lens; i++ {
			id = addr[i].ID
			if addr[i].Address == mainAddr {
				continue
			}
			wgcollect.Add(1)
			task <- true
			go func(k int) {
				collectall(addr[k].Address)
				wgcollect.Done()
				<-task
			}(i)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		time.Sleep(time.Second)
	}
}
