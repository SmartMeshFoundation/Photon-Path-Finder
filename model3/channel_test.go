package model3

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"github.com/stretchr/testify/assert"

	"github.com/SmartMeshFoundation/Photon/utils"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func setupdb() {
	os.Remove("test.db")
	SetUpDB("sqlite3", "test.db")
}
func TestSetUpDB(t *testing.T) {
	setupdb()
}

func TestChannel(t *testing.T) {
	setupdb()
	c := &Channel{
		ChannelID: utils.NewRandomHash().String(),
		Status:    ChannelStatusClosed,
		Token:     utils.NewRandomAddress().String(),
	}
	c.Participants = make([]*ChannelParticipantInfo, 2)
	c.Participants[0] = &ChannelParticipantInfo{
		Nonce: 1,
	}
	c.Participants[1] = &ChannelParticipantInfo{
		Nonce: 2,
	}
	if err := db.Create(c).Error; err != nil {
		t.Errorf("new channel error %s", err)
		return
	}
	if err := db.Create(c); err == nil {
		t.Error("cannot duplicate")
		return
	}
	var c2 Channel
	err := db.First(&c2, &Channel{ChannelID: c.ChannelID}).Error
	if err != nil {
		t.Error(err)
		return
	}
	if c.ChannelID != c2.ChannelID {
		t.Error("not equal")
		return
	}
	c3 := &Channel{
		ChannelID: c.ChannelID,
	}
	if err := db.Where(c3).First(c3).Error; err != nil {
		t.Error(err)
		return
	}
	if c3.ChannelID != c.ChannelID {
		t.Error("not equal")
		return
	}
	var c4 Channel
	if err := db.Debug().Where(c3).Preload("Participants").Find(&c4).Error; err != nil {
		t.Error(err)
		return
	}
	//db.Preloads("Participants").Find(c2)
	if c4.Participants == nil {
		//t.Logf("c=%s\nc4=%s", utils.StringInterface(c, 3), utils.StringInterface(c2, 3))
		t.Error("must equal")
	}
	t.Logf("c4=%s", utils.StringInterface(c4, 3))
}
func testCreateChannel(t *testing.T) (c2 *Channel) {
	token := utils.NewRandomAddress()
	channelIdentifier := utils.NewRandomHash()
	c := &Channel{
		ChannelID: channelIdentifier.String(),
		Status:    ChannelStatusOpen,
		Token:     token.String(),
	}
	participant1 := utils.NewRandomAddress()
	participant2 := utils.NewRandomAddress()
	c.Participants = make([]*ChannelParticipantInfo, 2)
	c.Participants[0] = &ChannelParticipantInfo{
		Participant: participant1.String(),
	}
	c.Participants[1] = &ChannelParticipantInfo{
		Participant: participant2.String(),
	}
	err := AddChannel(token, participant1, participant2, channelIdentifier, 3)
	if err != nil {
		t.Error(err)
		return
	}
	c2, err = getChannel(c.ChannelID)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	return c2
}
func TestAddChannel(t *testing.T) {
	setupdb()
	token := utils.NewRandomAddress()
	channelIdentifier := utils.NewRandomHash()
	c := &Channel{
		ChannelID: channelIdentifier.String(),
		Status:    ChannelStatusOpen,
		Token:     token.String(),
	}
	participant1 := utils.NewRandomAddress()
	participant2 := utils.NewRandomAddress()
	c.Participants = make([]*ChannelParticipantInfo, 2)
	c.Participants[0] = &ChannelParticipantInfo{
		Participant: participant1.String(),
	}
	c.Participants[1] = &ChannelParticipantInfo{
		Participant: participant2.String(),
	}
	err := AddChannel(token, participant1, participant2, channelIdentifier, 3)
	if err != nil {
		t.Error(err)
		return
	}
	c2, err := getChannel(c.ChannelID)
	assert.EqualValues(t, c.ChannelID, c2.ChannelID)
	assert.EqualValues(t, c.Status, c2.Status)
	assert.EqualValues(t, c.Token, c2.Token)
	assert.EqualValues(t, c.Participants[0].Participant, c2.Participants[0].Participant)
	assert.EqualValues(t, c.Participants[0].Nonce, c2.Participants[0].Nonce)
	assert.EqualValues(t, c.Participants[1].Participant, c2.Participants[1].Participant)
	assert.EqualValues(t, c.Participants[1].Nonce, c2.Participants[1].Nonce)
	assert.EqualValues(t, c2.Participants[0].FeePolicy, params.DefaultFeePolicy)
	assert.EqualValues(t, c2.Participants[1].FeeConstantPart, params.DefaultFeeConstantPart.String())
	assert.EqualValues(t, c2.Participants[0].FeePercentPart, params.DefaultFeePercentPart)
}

func TestCloseChannel(t *testing.T) {
	setupdb()
	c := testCreateChannel(t)
	assert.EqualValues(t, c.Status, ChannelStatusOpen)
	err := CloseChannel(common.HexToHash(c.ChannelID))
	if err != nil {
		t.Error(err)
		return
	}
	c2, err := getChannel(c.ChannelID)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, c2.Status, ChannelStatusClosed)
}

func TestSettleChannel(t *testing.T) {
	setupdb()
	c := testCreateChannel(t)
	err := SettleChannel(common.HexToHash(c.ChannelID))
	if err != nil {
		t.Error(err)
		return
	}
	_, err = getChannel(c.ChannelID)
	if err == nil {
		t.Error("should deleted")
	}
}

func TestUpdateChannelDeposit(t *testing.T) {
	setupdb()
	c := testCreateChannel(t)
	channelIdentifier := common.HexToHash(c.ChannelID)
	p1 := c.Participants[0]
	p2 := c.Participants[1]
	err := UpdateChannelDeposit(channelIdentifier, common.HexToAddress(p1.Participant), big.NewInt(20))
	if err != nil {
		t.Error(err)
		return
	}
	c2, _ := getChannel(c.ChannelID)
	assert.EqualValues(t, c2.Participants[0].Balance, big.NewInt(20).String())
	err = UpdateChannelDeposit(channelIdentifier, common.HexToAddress(p2.Participant), big.NewInt(30))
	if err != nil {
		t.Error(err)
		return
	}
	c2, _ = getChannel(c.ChannelID)
	assert.EqualValues(t, c2.Participants[0].Balance, big.NewInt(20).String())
	assert.EqualValues(t, c2.Participants[1].Balance, big.NewInt(30).String())
}
func TestUpdateChannelBalanceProof(t *testing.T) {
	setupdb()
	c := testCreateChannel(t)
	participant := common.HexToAddress(c.Participants[0].Participant)
	partner := common.HexToAddress(c.Participants[1].Participant)
	err := UpdateChannelDeposit(common.HexToHash(c.ChannelID), participant, big.NewInt(50))
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateChannelDeposit(common.HexToHash(c.ChannelID), partner, big.NewInt(50))
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateChannelBalanceProof(participant, partner, big.NewInt(0), &BalanceProof{
		ChannelID:      common.HexToHash(c.ChannelID),
		TransferAmount: big.NewInt(32),
		Nonce:          1,
		LocksRoot:      utils.NewRandomHash(),
	})
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateChannelBalanceProof(participant, partner, big.NewInt(0), &BalanceProof{
		ChannelID:      common.HexToHash(c.ChannelID),
		TransferAmount: big.NewInt(32),
		Nonce:          0,
		LocksRoot:      utils.NewRandomHash(),
	})
	if err == nil {
		t.Error("should failed because of nonce")
		return
	}
	err = UpdateChannelBalanceProof(participant, partner, big.NewInt(0), &BalanceProof{
		ChannelID:      common.HexToHash(c.ChannelID),
		TransferAmount: big.NewInt(22),
		Nonce:          3,
		LocksRoot:      utils.NewRandomHash(),
	})
	if err == nil {
		t.Error("should failed because of transfer amount decrease")
		return
	}
	c2, _ := getChannel(c.ChannelID)
	assert.EqualValues(t, c2.Participants[0].Balance, big.NewInt(82).String())
	assert.EqualValues(t, c2.Participants[1].Balance, big.NewInt(18).String())
	err = UpdateChannelBalanceProof(partner, participant, big.NewInt(50), &BalanceProof{
		ChannelID:      common.HexToHash(c.ChannelID),
		TransferAmount: big.NewInt(10),
		Nonce:          1,
		LocksRoot:      utils.NewRandomHash(),
	})
	if err != nil {
		t.Error(err)
		return
	}
	c2, _ = getChannel(c.ChannelID)
	assert.EqualValues(t, c2.Participants[0].Balance, big.NewInt(22).String())
	assert.EqualValues(t, c2.Participants[0].LockedAmount, big.NewInt(50).String())
	assert.EqualValues(t, c2.Participants[1].Balance, big.NewInt(28).String())
}
