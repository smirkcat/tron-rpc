package util

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"time"
	"tron/common/crypto"
	"tron/core"

	"github.com/golang/protobuf/proto"
)

// SignTransaction 签名交易
func SignTransaction(transaction *core.Transaction, key *ecdsa.PrivateKey) error {
	transaction.GetRawData().Timestamp = time.Now().UnixNano() / 1000000
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		return err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	contractList := transaction.GetRawData().GetContract()
	for range contractList {
		signature, err := crypto.Sign(hash, key)
		if err != nil {
			return err
		}
		transaction.Signature = append(transaction.Signature, signature)
	}
	return nil
}
