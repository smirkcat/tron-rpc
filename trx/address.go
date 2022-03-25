package trx

import (
	"crypto/ecdsa"
	"errors"
	"tron/model"

	"github.com/smirkcat/hdwallet"
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

func NewPrivateKey() (int, *ecdsa.PrivateKey, error) {
	if IsMulti {
		return 0, nil, errors.New("not suppot new addr is_multi true ")
	}
	index := dbengine.GetAccountMaxIndex() + 1
	ac, err := hdwallet.NewPrivateKeyIndex(index)
	return index, ac, err
}
