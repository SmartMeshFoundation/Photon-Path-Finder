package model

import (
	"testing"
)

func TestInitChannelView(t *testing.T) {

}

/*func TestChannelView_UpdateCapacity(t *testing.T) {
	xdeposit := big.NewInt(10)
	xamount := big.NewInt(0)

	oldChannelView := &ChannelView{
		SelfAddress:       utils.NewRandomAddress(),
		PartnerAddress:    utils.NewRandomAddress(),
		Deposit:           xdeposit,
		TransferredAmount: xamount,
		ReceivedAmount:    big.NewInt(0),
		LockedAmount:      big.NewInt(0),
		RrelativeFee:      big.NewInt(0),
		Capacity:          xdeposit,
		Status:            StateUpdateBalance,
		ChannelID:         utils.NewRandomHash(),
		BalanceProofNonce: &PeerNonce{utils.NewRandomAddress(), utils.NewRandomHash(), 50},
	}
	fmt.Printf("test ChannelView: %s", utils.StringInterface(oldChannelView, 1))

	var nonce = 0
	var deposit = big.NewInt(0)
	var transferredAmount = big.NewInt(0)
	var receivedAmount = big.NewInt(0)
	var lockedAmount = big.NewInt(0)
	err := oldChannelView.UpdateCapacity(nonce, deposit, transferredAmount, receivedAmount, lockedAmount)
	if err != nil {
		fmt.Printf("TestChannelView_UpdateCapacity err=%s", err)
	}
}*/
