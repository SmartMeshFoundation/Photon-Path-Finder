package model

import (
	"context"
	"math/big"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

//ChannelView Unidirectional view of a bidirectional channel
type ChannelView struct {
	Token             common.Address
	SelfAddress       common.Address
	PartnerAddress    common.Address
	Deposit           *big.Int
	TransferredAmount *big.Int
	ReceivedAmount    *big.Int
	LockedAmount      *big.Int
	RrelativeFee      *big.Int
	Capacity          *big.Int
	Status            string
	ChannelID         common.Hash
	BalanceProofNonce *PeerNonce
	db                storage.Database
	ctx               context.Context
}

//PeerNonce peer's nonce in which channel
type PeerNonce struct {
	PeerAddress common.Address
	ChannelID   common.Hash
	Nonce       int
}

const (
	// StateChannelOpen event type of ChannelOpen
	StateChannelOpen = "open"
	// StateChannelDeposit event type of ChannelDeposit
	StateChannelDeposit = "deposit"
	// StateChannelWithdraw event type of ChannelWithdraw
	StateChannelWithdraw = "withdraw"
	// StateChannelClose event type of ChannelClose
	StateChannelClose = "close"
	// StateUpdateBalance event type of UpdateBalance
	StateUpdateBalance = "updatebalance"
)

//InitChannelView some channel view
func InitChannelView(token common.Address, channelID common.Hash, participant1, participant2 common.Address, deposit *big.Int, eventTag string, balanceProofNonce *PeerNonce, db *storage.Database) *ChannelView {
	cv := &ChannelView{
		Token:             token,
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
		BalanceProofNonce: balanceProofNonce,
		db:                *db,
	}
	return cv
}

//UpdateCapacity refush channel status and capacity
func (cv *ChannelView) UpdateCapacity(
	nonce uint64,
	deposit *big.Int,
	transferredAmount *big.Int,
	receivedAmount *big.Int,
	lockedAmount *big.Int,
) (err error) {
	switch cv.Status {
	case StateChannelOpen:
		err = cv.db.UpdateChannelStatusStorage(nil, cv.Token.String(), cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	case StateChannelDeposit:
		err = cv.db.UpdateChannelDepositStorage(nil, cv.Token.String(), cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String(), cv.Capacity.Int64())
	case StateChannelWithdraw:
		err = cv.db.WithdrawChannelInfoStorage(nil, cv.Token.String(), cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String(), deposit.Int64())
	case StateChannelClose:
		err = cv.db.UpdateChannelStatusStorage(nil, cv.Token.String(), cv.ChannelID.String(), cv.Status, cv.SelfAddress.String(), cv.PartnerAddress.String())
	case StateUpdateBalance:
		/*if nonce != 0 {
			if nonce <= cv.BalanceProofNonce.Nonce {
				logrus.Warn("Balance proof nonce must increase.")
				return
			}
			cv.BalanceProofNonce.Nonce = nonce
		}*/
		/*if deposit.Uint64() != 0 {
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
		cv.Capacity.Add(cv.Capacity, cv.ReceivedAmount)*/
		err = cv.db.UpdateBalanceProofStorage(nil, cv.Token.String(),
			cv.ChannelID.String(),
			cv.Status,
			cv.SelfAddress.String(),
			cv.PartnerAddress.String(),
			transferredAmount.Int64(), receivedAmount.Int64(), lockedAmount.Int64(),
			nonce,
		)
	}
	if err != nil {
		logrus.Warn("Failed to update capacity,err=", cv.Status)
	}
	return
}
