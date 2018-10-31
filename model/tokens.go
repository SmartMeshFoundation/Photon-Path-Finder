package model3

import (
	"fmt"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/ethereum/go-ethereum/common"
)

type tokenNetwork struct {
	Token        string `gorm:"primary_key"`
	TokenNetwork string `gorm:"primary_key"`
	BlockNumber  int64
}

type latestBlockNumber struct {
	ID          int
	BlockNumber int64
}

var lb = &latestBlockNumber{ID: 1}

//UpdateBlockNumber 更新最新的已经处理过的块数
func UpdateBlockNumber(blockNumber int64) {
	err := db.Model(lb).UpdateColumn("BlockNumber", blockNumber).Error
	if err != nil {
		log.Crit(fmt.Sprintf("err=%s", err))
	}
}

//GetLatestBlockNumber 获取已经处理的最新块数
func GetLatestBlockNumber() int64 {
	l2 := &latestBlockNumber{}
	if err := db.First(l2).Error; err != nil {
		log.Crit(fmt.Sprintf("err=%s", err))
	}
	return l2.BlockNumber
}

//AddTokeNetwork 链上新建了一个tokennetwork
func AddTokeNetwork(token, tw common.Address, blockNumber int64) error {
	t := &tokenNetwork{
		Token:        token.String(),
		TokenNetwork: tw.String(),
		BlockNumber:  blockNumber,
	}
	return db.Create(t).Error
}

//GetAllTokenNetworks 获取目前已知的tokenNetwork
func GetAllTokenNetworks() map[common.Address]common.Address {
	m := make(map[common.Address]common.Address)
	var ts []tokenNetwork
	if err := db.Find(&ts).Error; err != nil {
		log.Crit(err.Error())
	}
	for i := range ts {
		m[common.HexToAddress(ts[i].Token)] = common.HexToAddress(ts[i].TokenNetwork)
	}
	return m
}
