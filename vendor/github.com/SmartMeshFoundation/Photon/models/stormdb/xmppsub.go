package stormdb

import (
	"fmt"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/models"
	"github.com/ethereum/go-ethereum/common"
)

//XMPPMarkAddrSubed mark `addr` subscribed
func (model *StormDB) XMPPMarkAddrSubed(addr common.Address) {
	err := model.db.Set(models.BucketXMPP, addr[:], true)
	if err != nil {
		log.Error(fmt.Sprintf("db err %s", err))
	}
}

//XMPPIsAddrSubed return true when `addr` already subscirbed
func (model *StormDB) XMPPIsAddrSubed(addr common.Address) bool {
	var r bool
	err := model.db.Get(models.BucketXMPP, addr[:], &r)
	if err != nil {
		log.Trace(fmt.Sprintf("db err %s", err))
	}
	return r
}

//XMPPUnMarkAddr mark `addr` has been unsubscribed
func (model *StormDB) XMPPUnMarkAddr(addr common.Address) {
	err := model.db.Set(models.BucketXMPP, addr[:], false)
	if err != nil {
		log.Error(fmt.Sprintf("db err %s", err))
	}
}
