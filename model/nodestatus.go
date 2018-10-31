package model

import (
	"fmt"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/ethereum/go-ethereum/common"
)

type NodeStatus struct {
	Address    string `gorm:"primary_key"`
	DeviceType string
	IsOnline   bool
}

func GetAllNodes() []*NodeStatus {
	var nodes []*NodeStatus

	if err := db.Limit(1000000).Find(&nodes).Error; err != nil {
		log.Crit(err.Error())
	}
	return nodes
}

func NewOrUpdateNodeOnline(address common.Address, isOnline bool) {
	var node = &NodeStatus{}
	node.Address = address.String()
	if err := db.Where(node).Find(node).Error; err != nil {
		log.Error(fmt.Sprintf("NewOrUpdateNodeOnline but cannot found node %s", address.String()))
		//return
		db.Create(node)
	}
	node.IsOnline = isOnline
	err := db.Model(node).UpdateColumn("IsOnline", isOnline).Error
	if err != nil {
		log.Error(fmt.Sprintf("update online err %s", err))
	}
}
func NewOrUpdateNodeStatus(address common.Address, isOnline bool, deviceType string) {
	var node = &NodeStatus{}
	node.Address = address.String()
	if err := db.Where(node).Find(node).Error; err != nil {
		//log.Error(fmt.Sprintf("NewOrUpdateNodeStatus but cannot found node %s", address.String()))
		//return
		err = db.Create(node).Error
		if err != nil {
			log.Error("create error %s", err)
		}
	}
	node.IsOnline = isOnline
	err := db.Model(node).UpdateColumns(&NodeStatus{
		IsOnline:   isOnline,
		DeviceType: deviceType,
	}).Error
	if err != nil {
		log.Error(fmt.Sprintf("update online err %s", err))
	}
}
