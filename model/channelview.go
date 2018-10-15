package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"context"
	"fmt"
	"math/big"
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
)

//InitChannelView
func InitChannelView(channelID common.Hash ,participant1,participant2 common.Address,deposit *big.Int,eventTag string) (*ChannelView) {
	cv := &ChannelView{
		SelfAddress:       participant1,
		PartnerAddress:    participant2,
		Deposit:           deposit,
		TransferredAmount: big.NewInt(0),
		ReceivedAmount:    big.NewInt(0),
		LockedAmount:      big.NewInt(0),
		RrelativeFee:      big.NewInt(0),
		Capacity:          deposit,
		Status:             eventTag,
		ChannelID:         channelID,
		BalanceProofNonce: 0,
	}
	return cv
}

//UpdateCapacity
func (cv ChannelView)UpdateCapacity(
	nonce int64,
	deposit *big.Int,
	transferredAmount *big.Int,
	receivedAmount *big.Int,
	lockedAmount *big.Int,
	)(err error) {
	switch cv.Status {
	case StateChannelOpen:
		err = cv.db.CreateChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	case StateChannelDeposit:
		if nonce != 0 && nonce > cv.BalanceProofNonce {
			cv.BalanceProofNonce = nonce
		}
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
		//cv.Capacity=cv.Deposit. - (cv.TransferredAmount + cv.LockedAmount) + cv.ReceivedAmount
		cv.Capacity.Sub(cv.Capacity, cv.TransferredAmount)
		cv.Capacity.Sub(cv.Capacity, cv.LockedAmount)
		cv.Capacity.Add(cv.Capacity, cv.ReceivedAmount)

		err = cv.db.UpdateChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.Capacity)
	case StateChannelWithdraw:
		err = cv.db.WithdrawChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), deposit)
	case StateChannelClose:
		err = cv.db.CreateChannelInfoStorage(nil, cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	}
	//Refresh memory
	//..
	if err != nil {
		fmt.Println("failed to update capacity")
	}
	return
}
