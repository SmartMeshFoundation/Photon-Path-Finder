package model

import (
	"testing"

	"github.com/SmartMeshFoundation/Photon/utils"
)

func TestGetLatestBlockNumber(t *testing.T) {
	SetupTestDB()
	if GetLatestBlockNumber() != 0 {
		t.Error("should be 0 first")
		return
	}
	UpdateBlockNumber(32)
	if GetLatestBlockNumber() != 32 {
		t.Error("should 32")
		return
	}
}

func TestGetAllTokenNetworks(t *testing.T) {
	SetupTestDB()
	a1 := utils.NewRandomAddress()
	a2 := utils.NewRandomAddress()
	m := GetAllTokenNetworks()
	err := AddTokeNetwork(a1, a2, 3)
	if err != nil {
		t.Error(err)
		return
	}
	err = AddTokeNetwork(a1, a2, 5)
	if err == nil {
		t.Error("cannot duplicate")
		return
	}
	m = GetAllTokenNetworks()
	if len(m) != 1 {
		t.Error("length error")
	}
}
