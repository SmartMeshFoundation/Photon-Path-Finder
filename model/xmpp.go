package model

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/ethereum/go-ethereum/common"
)

type xmpp struct {
	gorm.Model
	Address      string `gorm:"primary_key"` //地址作为唯一的key,因为在整个网络中我只需关注
	IsSubScribed bool   //是否已经订阅改地址
}

func getxmpp(address string) (x *xmpp, err error) {
	x = &xmpp{
		Address: address,
	}
	err = db.Where(x).Find(x).Error
	return
}

//XMPPMarkAddrSubed mark `addr` subscribed
func XMPPMarkAddrSubed(addr common.Address) {
	x, err := getxmpp(addr.String())
	if err == nil {
		x.IsSubScribed = true
		err = db.Model(x).UpdateColumn("IsSubScribed", true).Error
	} else {
		x := &xmpp{
			Address:      addr.String(),
			IsSubScribed: true,
		}
		err = db.Create(x).Error
	}
	if err != nil {
		log.Error(fmt.Sprintf("XMPPMarkAddrSubed %s err %s", addr.String(), err))
	}
}

//XMPPIsAddrSubed return true when `addr` already subscirbed
func XMPPIsAddrSubed(addr common.Address) bool {
	x, err := getxmpp(addr.String())
	return err == nil && x.IsSubScribed
}

//XMPPUnMarkAddr mark `addr` has been unsubscribed
func XMPPUnMarkAddr(addr common.Address) {
	x, err := getxmpp(addr.String())
	if err == nil {
		x.IsSubScribed = false
		err = db.Model(x).UpdateColumn("IsSubScribed", false).Error
	}
	if err != nil {
		log.Error(fmt.Sprintf("XMPPUnMarkAddr %s err %s", addr.String(), err))
	}
}
