package trx

import (
	"tron/model"
)

func SearchAccount(addr string) (*Account, error) {
	var ac *Account
	var err error
	if IsMulti {
		acmuti, err := model.GetUsedOfAddressMulti("tron", addr)
		if err != nil || acmuti.Id < 1 {
			return nil, err
		}
		ac = &Account{
			Address:    acmuti.TronAddress,
			PublicKey:  acmuti.PublicKey,
			PrivateKey: acmuti.PrivateKey,
			User:       acmuti.Account,
		}
		return ac, nil
	}
	ac, err = dbengine.GetAccountWithAddr(addr)
	return ac, err
}

func InitAddressDB(dsn string) {
	if IsMulti {
		model.InitDB(dsn)
	}
}
