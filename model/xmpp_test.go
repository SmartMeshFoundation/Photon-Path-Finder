package model

import (
	"testing"

	"github.com/SmartMeshFoundation/Photon/utils"
)

func TestXMPPIsAddrSubed(t *testing.T) {
	SetupTestDB()
	addr1 := utils.NewRandomAddress()
	if XMPPIsAddrSubed(addr1) {
		t.Error("should not sub")
		return
	}
	//should not error
	XMPPUnMarkAddr(addr1)
	//no error
	XMPPMarkAddrSubed(addr1)
	if !XMPPIsAddrSubed(addr1) {
		t.Error("should subed")
		return
	}
	XMPPUnMarkAddr(addr1)
	if XMPPIsAddrSubed(addr1) {
		t.Error("should not subed")
	}
}
