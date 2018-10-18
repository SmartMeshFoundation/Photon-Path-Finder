package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"context"
	"math/big"
	"github.com/sirupsen/logrus"
)

//ChannelView Unidirectional view of a bidirectional channel
type ChannelView struct {
	SelfAddress       common.Address
	PartnerAddress    common.Address
	Deposit           *big.Int
	TransferredAmount *big.Int
	ReceivedAmount    *big.Int
	LockedAmount      *big.Int
	RrelativeFee      *big.Int
	Capacity          *big.Int
	Status             string
	ChannelID         common.Hash
	BalanceProofNonce int64
	db                storage.Database
	ctx               context.Context
}

const(
	//StateChannelOpen event type of ChannelOpen
	StateChannelOpen  ="open"
	//StateChannelDeposit event type of ChannelDeposit
	StateChannelDeposit="deposit"
	///StateChannelWithdraw event type of ChannelWithdraw
	StateChannelWithdraw="withdraw"
	//StateChannelOpen event type of ChannelClose
	StateChannelClose="close"
	//StateChannelOpen event type of ChannelClose
	StateUpdateBalance="updatebalance"
)

//InitChannelView
func InitChannelView(channelID common.Hash ,participant1,participant2 common.Address,deposit *big.Int,eventTag string,db *storage.Database) (*ChannelView) {
	cv := &ChannelView{
		SelfAddress:       participant1,
		PartnerAddress:    participant2,
		Deposit:           deposit,
		TransferredAmount: big.NewInt(0),
		ReceivedAmount:    big.NewInt(0),
		LockedAmount:      big.NewInt(0),
		RrelativeFee:      big.NewInt(0),
		Capacity:          deposit,
		Status:            eventTag,
		ChannelID:         channelID,
		BalanceProofNonce: 0,
		db:                *db,
	}
	return cv
}

//UpdateCapacity refush channel status and capacity
func (cv *ChannelView)UpdateCapacity(
	nonce int64,
	deposit *big.Int,
	transferredAmount *big.Int,
	receivedAmount *big.Int,
	lockedAmount *big.Int,
	)(err error) {
	switch cv.Status {
	case StateChannelOpen:
		err = cv.db.UpdateChannelStatusStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	case StateChannelDeposit:
		/*if nonce != 0 && nonce > cv.BalanceProofNonce {
			cv.BalanceProofNonce = nonce
		} else {
			logrus.Warn("Balance proof nonce must increase.")
			return
		}*/
		if deposit.Uint64() != 0 {
			cv.Deposit = deposit
		}
		if transferredAmount.Uint64() != 0 {
			cv.TransferredAmount = transferredAmount
		}
		if receivedAmount.Uint64() != 0 {
			cv.ReceivedAmount = receivedAmount
		}
		if lockedAmount.Uint64() != 0 {
			cv.LockedAmount = lockedAmount
		}
		cv.Capacity.Sub(cv.Capacity, cv.TransferredAmount)
		cv.Capacity.Sub(cv.Capacity, cv.LockedAmount)
		cv.Capacity.Add(cv.Capacity, cv.ReceivedAmount)
		err = cv.db.UpdateChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String(), cv.Capacity.Int64())
	case StateChannelWithdraw:
		err = cv.db.WithdrawChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String(), deposit.Int64())
	case StateChannelClose:
		err = cv.db.UpdateChannelStatusStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	case StateUpdateBalance:
		if deposit.Uint64() != 0 {
			cv.Deposit = deposit
		}
		if transferredAmount.Uint64() != 0 {
			cv.TransferredAmount = transferredAmount
		}
		if receivedAmount.Uint64() != 0 {
			cv.ReceivedAmount = receivedAmount
		}
		if lockedAmount.Uint64() != 0 {
			cv.LockedAmount = lockedAmount
		}
		cv.Capacity.Sub(cv.Capacity, cv.TransferredAmount)
		cv.Capacity.Sub(cv.Capacity, cv.LockedAmount)
		cv.Capacity.Add(cv.Capacity, cv.ReceivedAmount)
		err = cv.db.UpdateBalanceProofStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String(), cv.Capacity.Int64())
	}
	if err != nil {
		logrus.Warn("Failed to update capacity,err=", err)
	}
	return
}