package model

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"github.com/SmartMeshFoundation/Photon/utils"
)

func TestGetAccountFeePolicy(t *testing.T) {
	SetupTestDB()
	a := utils.NewRandomAddress()
	fee := GetAccountFeePolicy(a)
	if fee.FeePolicy != params.DefaultFeePolicy ||
		fee.FeePercent != params.DefaultFeePercentPart {
		t.Error("not equal default")
	}
	fee.FeePolicy = FeePolicyConstant
	fee.FeePercent = 0
	fee.FeeConstant = big.NewInt(30)
	err := UpdateAccountDefaultFeePolicy(a, fee)
	if err != nil {
		t.Error(err)
		return
	}
	fee2 := GetAccountFeePolicy(a)
	if !reflect.DeepEqual(fee2, fee) {
		t.Error("not equal")
		return
	}
}

func TestGetAccountTokenFee(t *testing.T) {
	SetupTestDB()
	a := utils.NewRandomAddress()
	token := utils.NewRandomAddress()
	fee, err := GetAccountTokenFee(a, token)
	if err == nil {
		t.Error("should not found")
		return
	}
	fee = &Fee{}
	fee.FeePolicy = FeePolicyConstant
	fee.FeePercent = 0
	fee.FeeConstant = big.NewInt(30)
	err = UpdateAccountTokenFee(a, token, fee)
	if err != nil {
		t.Error(err)
		return
	}
	fee2, err := GetAccountTokenFee(a, token)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(fee2, fee) {
		t.Error("not equal")
		return
	}
}
