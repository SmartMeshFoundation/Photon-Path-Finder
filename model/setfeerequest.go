package model

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// SetFeeRateRequest is the json request for setChannelRate
type SetFeeRateRequest struct {
	FeeConstant *big.Int `json:"fee_constant"`
	FeePercent  int64    `json:"fee_percent"`
	Signature   []byte   `json:"signature"`
	Fee         *Fee     `json:"-"`
}

//SetAllFeeRateRequest is json request for all fee rate
type SetAllFeeRateRequest struct {
	AccountFee  *SetFeeRateRequest                    `json:"account_fee"`
	TokensFee   map[common.Address]*SetFeeRateRequest `json:"token_fee_map"`
	ChannelsFee map[common.Hash]*SetFeeRateRequest    `json:"channel_fee_map"`
}
