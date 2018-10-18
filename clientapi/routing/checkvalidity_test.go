package routing

import (
	"testing"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"math/big"
)

func TestVerifySinature(t *testing.T) {
	br := &BalanceProof{
		Nonce:             1,
		TransferredAmount: big.NewInt(32),
		LocksRoot:         utils.EmptyHash,
		AdditionalHash:    utils.EmptyHash,
		ChannelID:         utils.NewRandomHash(),
	}

	key1, addr1 := utils.MakePrivateKeyAddress()
	key2, addr2 := utils.MakePrivateKeyAddress()
	err := SignDataForBalanceProof0(key1, br)
	if err != nil {
		t.Error(err)
		return
	}
	brm := &balanceProofRequest{
		BalanceProof: *br,
		LocksAmount:  big.NewInt(0),
	}
	err = SignDataForBalanceProofMessage0(key2, brm)
	if err != nil {
		t.Error(err)
		return
	}

	err = verifySinature(brm, addr2, addr1)
	if err != nil {
		t.Error(err)
		return
	}
}