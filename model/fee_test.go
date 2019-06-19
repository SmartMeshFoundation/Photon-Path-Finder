package model

import (
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

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
	fee.FeePercent = 30
	err = UpdateAccountDefaultFeePolicy(a, fee)
	if err != nil {
		t.Error(err)
		return
	}
	fee2 := GetAccountFeePolicy(a)
	if !reflect.DeepEqual(fee2, fee) {
		t.Error("not equal")
		return
	}
	err=DeleteAccountAllFeeRate(a)
	if err!=nil{
		t.Error(err)
		return
	}
	//删除后应该是缺省的
	fee = GetAccountFeePolicy(a)
	if fee.FeePolicy != params.DefaultFeePolicy ||
		fee.FeePercent != params.DefaultFeePercentPart {
		t.Error("not equal default")
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
	err = UpdateAccountTokenFee(a, token, fee)
	if err != nil {
		t.Error(err)
		return
	}
	err=DeleteAccountAllFeeRate(a)
	if err!=nil{
		t.Error(err)
		return
	}
	//删除后应该是缺省的
	fee,err = GetAccountTokenFee(a,token)
	if err==nil{
		t.Error("must not found")
		return
	}
}

func TestGetAccountTokenFeeSqlite(t *testing.T) {
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
	wg:=sync.WaitGroup{
	}
	n:=100
	wg.Add(n)
	start:=time.Now()
	for i:=0;i<n;i++{
		go func(){
			err = UpdateAccountTokenFee(a, token, fee)
			if err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	t.Logf("write %d cost %s",n,time.Since(start))
}
