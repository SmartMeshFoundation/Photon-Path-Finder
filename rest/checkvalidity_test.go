package rest

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon/utils"
)

func TestVerifySinature(t *testing.T) {
	br := &model.BalanceProof{
		Nonce:          1,
		TransferAmount: big.NewInt(32),
		LocksRoot:      utils.EmptyHash,
		AdditionalHash: utils.EmptyHash,
		ChannelID:      utils.NewRandomHash(),
	}

	key1, addr1 := utils.MakePrivateKeyAddress()
	key2, addr2 := utils.MakePrivateKeyAddress()
	err := SignDataForBalanceProof0(key1, br)
	if err != nil {
		t.Error(err)
		return
	}
	brm := &balanceProofRequest{
		BalanceProof: br,
		LockedAmount: big.NewInt(0),
	}
	err = SignDataForBalanceProofMessage0(key2, brm)
	if err != nil {
		t.Error(err)
		return
	}

	maddr, err := verifyBalanceProofSignature(brm, addr2)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, maddr, addr1)
}
