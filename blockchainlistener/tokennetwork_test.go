package blockchainlistener

import (
	"math"
	"math/big"
	"testing"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
)

type channelIDStruct struct {
	p1           common.Address
	p2           common.Address
	tokenNetwork common.Address
	channelID    common.Hash
}

func TestCalcChannelID(t *testing.T) {
	model.SetupTestDB()
	cases := []channelIDStruct{
		{
			p1:           common.HexToAddress("0x292650fee408320D888e06ed89D938294Ea42f99"),
			p2:           common.HexToAddress("0x192650FEe408320D888E06Ed89D938294EA42f99"),
			tokenNetwork: common.HexToAddress("0x6021334197e07966330BEd0dB7561a2EC5DC9A8A"),
			channelID:    common.HexToHash("0xd8b6510752125b1c3b826bfe730f3dc280792fad7c8c1d95415f468da955a154"),
		},
		{
			p1:           common.HexToAddress("0x292650fee408320D888e06ed89D938294Ea42f99"),
			p2:           common.HexToAddress("0x4B89Bff01009928784eB7e7d10Bf773e6D166066"),
			tokenNetwork: common.HexToAddress("0x6021334197e07966330BEd0dB7561a2EC5DC9A8A"),
			channelID:    common.HexToHash("0x12b4e8dd0d831a92de199b6b814861547b3109e2155841a673475053a42f8306"),
		},
	}
	for _, c := range cases {
		cid := calcChannelID(c.tokenNetwork, c.p1, c.p2)
		assert.EqualValues(t, cid, c.channelID)
		cid = calcChannelID(c.tokenNetwork, c.p2, c.p1)
		assert.EqualValues(t, cid, c.channelID)
	}
}

func TestTokenNetwork_GetPaths(t *testing.T) {
	model.SetupTestDB()
	tn := NewTokenNetwork(nil)
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn.decimals = map[common.Address]int{
		token: 0,
	}
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	tn.tokenNetwork2Token = map[common.Address]common.Address{
		tokenNetwork: token,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	c1 := &channel{
		Participant1: addr1,
		Participant2: addr2,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c1Id := calcChannelID(tokenNetwork, addr1, addr2)
	tn.channelViews[token] = []*channel{c1}
	tn.channels[c1Id] = c1
	paths, err := tn.GetPaths(addr1, addr2, token, big.NewInt(10), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr2, token, big.NewInt(30), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}
	c2 := &channel{
		Participant1: addr2,
		Participant2: addr3,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c2Id := calcChannelID(tokenNetwork, addr2, addr3)
	tn.channelViews[token] = []*channel{c1, c2}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(30), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}
}

func TestTokenNetwork_GetPathsBigInt(t *testing.T) {
	tn := NewTokenNetwork(nil)
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn.decimals = map[common.Address]int{
		token: 18,
	}
	base := big.NewInt(int64(math.Pow10(18)))
	balance := big.NewInt(20)
	balance = balance.Mul(balance, base)
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	tn.tokenNetwork2Token = map[common.Address]common.Address{
		tokenNetwork: token,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	fee := big.NewInt(1)
	fee.Mul(fee, base)
	c1 := &channel{
		Participant1: addr1,
		Participant2: addr2,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: fee,
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: fee,
		},
		Participant1Balance: balance,
		Participant2Balance: balance,
	}
	c1Id := calcChannelID(tokenNetwork, addr1, addr2)
	tn.channelViews[token] = []*channel{c1}
	tn.channels[c1Id] = c1
	v := big.NewInt(10)
	paths, err := tn.GetPaths(addr1, addr2, token, v.Mul(v, base), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	v = big.NewInt(30)
	paths, err = tn.GetPaths(addr1, addr2, token, v.Mul(v, base), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}
	c2 := &channel{
		Participant1: addr2,
		Participant2: addr3,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyCombined,
			FeeConstant: fee,
			FeePercent:  1000, //额外收千分之一
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: fee,
		},
		Participant1Balance: balance,
		Participant2Balance: balance,
	}
	c2Id := calcChannelID(tokenNetwork, addr2, addr3)
	tn.channelViews[token] = []*channel{c1, c2}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	v = big.NewInt(3)
	paths, err = tn.GetPaths(addr1, addr3, token, v.Mul(v, base), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	t.Logf("paths=%s", utils.StringInterface(paths, 3))
	v = big.NewInt(30)
	paths, err = tn.GetPaths(addr1, addr3, token, v.Mul(v, base), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}
}
