package model

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type observerKey struct {
	ID  int
	Key []byte
}

//GetObserverKey get or create a key from db
func GetObserverKey() *ecdsa.PrivateKey {
	x := &observerKey{
		ID: 1,
	}
	err := db.Where(x).Find(x).Error
	if err != nil {
		//第一次启动
		key, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		x.Key = crypto.FromECDSA(key)
		err = db.Create(x).Error
		if err != nil {
			panic(err)
		}
	}
	return crypto.ToECDSAUnsafe(x.Key)
}
