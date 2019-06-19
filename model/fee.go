package model

import (
	"math/big"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
)

//AccountFee 账户的缺省收费
type AccountFee struct {
	Account         string `gorm:"primary_key"`
	FeePolicy       int
	FeeConstantPart string
	FeePercentPart  int64
}

// AccountTokenFee 某个账户针对某个Token的缺省收费
type AccountTokenFee struct {
	gorm.Model
	Token           string `gorm:"index"` //调用者保证Token+Account必须是唯一的
	Account         string `gorm:"index"`
	FeePolicy       int
	FeeConstantPart string
	FeePercentPart  int64
}

//TokenFee 针对某种token 的缺省收费,暂不启用
type TokenFee struct {
	Token           string `gorm:"primary_key"`
	FeePolicy       int
	FeeConstantPart string
	FeePercentPart  int64
}

//Fee 为了使用方便定义
type Fee struct {
	FeePolicy   int      `json:"fee_policy"`
	FeeConstant *big.Int `json:"fee_constant" `
	FeePercent  int64    `json:"fee_percent"`
}

//UpdateAccountDefaultFeePolicy 设置某个账户的缺省收费,新创建的通道都会按照此缺省设置进行
func UpdateAccountDefaultFeePolicy(account common.Address, fee *Fee) error {
	a := &AccountFee{
		Account:         account.String(),
		FeePolicy:       fee.FeePolicy,
		FeeConstantPart: bigIntToString(fee.FeeConstant),
		FeePercentPart:  fee.FeePercent,
	}
	err := db.Where(&AccountFee{Account: account.String()}).Find(&AccountFee{}).Error
	if err == nil {
		return db.Save(a).Error
	}
	return db.Create(a).Error
}

var defaultFee = &Fee{
	FeePolicy:   params.DefaultFeePolicy,
	FeeConstant: params.DefaultFeeConstantPart,
	FeePercent:  params.DefaultFeePercentPart,
}

//GetAccountFeePolicy 获取某个账户的缺省收费,新创建的通道都会按照此缺省设置进行
func GetAccountFeePolicy(account common.Address) (fee *Fee) {
	a := &AccountFee{}
	err := db.Where(&AccountFee{Account: account.String()}).Find(a).Error
	if err == nil {
		return &Fee{
			FeePolicy:   a.FeePolicy,
			FeeConstant: stringToBigInt(a.FeeConstantPart),
			FeePercent:  a.FeePercentPart,
		}
	}
	return &Fee{
		defaultFee.FeePolicy, defaultFee.FeeConstant, defaultFee.FeePercent,
	}
}

// GetAccountTokenFee 获取账户针对某个token的缺省收费设置
func GetAccountTokenFee(account, token common.Address) (fee *Fee, err error) {
	atf := &AccountTokenFee{
		Token:   token.String(),
		Account: account.String(),
	}
	err = db.Where(atf).Find(atf).Error
	if err == nil {
		fee = &Fee{
			FeePolicy:   atf.FeePolicy,
			FeeConstant: stringToBigInt(atf.FeeConstantPart),
			FeePercent:  atf.FeePercentPart,
		}
	}
	return
}

//UpdateAccountTokenFee 更新用户针对某个token的缺省收费设置
func UpdateAccountTokenFee(account, token common.Address, fee *Fee) (err error) {
	atf := &AccountTokenFee{
		Token:   token.String(),
		Account: account.String(),
	}
	err = db.Where(atf).Find(atf).Error
	atf.FeePolicy = fee.FeePolicy
	atf.FeeConstantPart = bigIntToString(fee.FeeConstant)
	atf.FeePercentPart = fee.FeePercent
	if err == nil {
		return db.Save(atf).Error
	}
	return db.Create(atf).Error
}
//DeleteAccountAllFeeRate 删除账户所有收费记录
func DeleteAccountAllFeeRate(account common.Address) (err error){
	tx:=db.Begin()
	if err!=nil{
		tx.Rollback()
	}
	tx2:=tx.Where("account=?",account.String()).Delete(&AccountFee{})
	if tx2.Error!=nil{
		err=tx2.Error
		return
	}
	tx2=tx.Where("account=?",account.String()).Delete(&AccountTokenFee{})
	if tx2.Error!=nil{
		err=tx2.Error
		return
	}
	tx2=tx.Where("participant=?",account.String()).Delete(&ChannelParticipantFee{})
	if tx2.Error!=nil{
		err=tx2.Error
		return
	}
	return tx.Commit().Error
}