package model

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
)

type observerKey struct {
	gorm.Model
	Key []byte
}

//GetObserverKey get or create a key from db
func GetObserverKey() *ecdsa.PrivateKey {
	x := &observerKey{
		Model: gorm.Model{ID: 1},
	}
	err := db.Where(x).Find(x).Error
	if err != nil {
		//第一次启动
		key, _ := crypto.GenerateKey()
		x.Key = crypto.FromECDSA(key)
		err = db.Create(x).Error
		if err != nil {
			panic(err)
		}
	}
	return crypto.ToECDSAUnsafe(x.Key)
}
