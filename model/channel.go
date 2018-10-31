package model3

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	// import _ "github.com/jinzhu/gorm/dialects/postgres"
	// import _ "github.com/jinzhu/gorm/dialects/sqlite"
	// import _ "github.com/jinzhu/gorm/dialects/mssql"
)

const (
	//ChannelStatusOpen 通道状态打开
	ChannelStatusOpen = iota
	//ChannelStatusClosed 通道状态关闭
	ChannelStatusClosed
	//ChannelStatusSettled 通道结算状态
	ChannelStatusSettled
)

//SettledChannel 在数据库中存储已经结算的通道,为了以后使用
type SettledChannel struct {
	gorm.Model
	ChannelID    string
	Participant1 string
	Participant2 string
	Data         string
}

const (
	//FeePolicyConstant 每笔交易,不论金额,固定收费
	FeePolicyConstant = iota
	//FeePolicyPercent 每笔交易固定收取一定比例的费用
	FeePolicyPercent
	//FeePolicyCombined 以上两种方式的组合
	FeePolicyCombined
)

/*ChannelParticipantInfo 通道中的一方需要存储的交易信息
由于数据库存储限制,
*/
type ChannelParticipantInfo struct {
	ID               int
	ChannelID        string `gorm:"index"`
	Participant      string
	Nonce            uint64
	Balance          string
	Deposit          string
	LockedAmount     string
	TransferedAmount string
	FeePolicy        int
	FeeConstantPart  string //固定部分是一个整数,比如一次收取1token
	FeePercentPart   int64  //0表示不收费,1000表示收费千分之一
}

//Channel Channel基本信息
type Channel struct {
	ChannelID       string `gorm:"primary_key"`
	Token           string `gorm:"index"`
	OpenBlockNumber int64
	Status          int
	Participants    []*ChannelParticipantInfo `gorm:"ForeignKey:ChannelID"`
}

func getChannel(channelID string) (c *Channel, err error) {
	c = &Channel{
		ChannelID: channelID,
	}
	err = db.Where(c).Preload("Participants").Find(c).Error
	return
}

//AddChannel add channel to db
func AddChannel(token, participant1, participant2 common.Address, ChannelIdentifier common.Hash, blockNumber int64) error {
	channelID := ChannelIdentifier.String()
	c, err := getChannel(channelID)
	if err == nil {
		return fmt.Errorf("channelId %s duplicate", channelID)
	}
	c = &Channel{ChannelID: channelID}
	c.Token = token.String()
	c.Status = ChannelStatusOpen
	c.OpenBlockNumber = blockNumber
	p1 := &ChannelParticipantInfo{
		Participant: participant1.String(),
	}
	fee, err := GetAccountTokenFee(participant1, token)
	if err != nil {
		fee = GetAccountFeePolicy(participant1)
	}
	p1.FeePercentPart = fee.FeePercent
	p1.FeeConstantPart = bigIntToString(fee.FeeConstant)
	p1.FeePolicy = fee.FeePolicy
	p2 := &ChannelParticipantInfo{
		Participant: participant2.String(),
	}
	fee, err = GetAccountTokenFee(participant1, token)
	if err != nil {
		fee = GetAccountFeePolicy(participant1)
	}
	p2.FeePercentPart = fee.FeePercent
	p2.FeeConstantPart = bigIntToString(fee.FeeConstant)
	p2.FeePolicy = fee.FeePolicy
	c.Participants = []*ChannelParticipantInfo{p1, p2}
	return db.Create(c).Error
}

// BalanceProof is the json request for BalanceProof
type BalanceProof struct {
	Nonce           uint64      `json:"nonce"`
	TransferAmount  *big.Int    `json:"transfer_amount"`
	LocksRoot       common.Hash `json:"locks_root"`
	ChannelID       common.Hash `json:"channel_identifier"`
	OpenBlockNumber int64       `json:"open_block_number"`
	AdditionalHash  common.Hash `json:"addition_hash"`
	Signature       []byte      `json:"signature"`
	ExtraHash       common.Hash `json:"extra_hash"`
}

func verifyParticipants(c *Channel, participant1, participant2 common.Address) (p1, p2 *ChannelParticipantInfo, err error) {
	p1 = c.Participants[0]
	p2 = c.Participants[1]

	if participant1.String() == p1.Participant && participant2.String() == p2.Participant {

	} else if participant1.String() == p2.Participant && participant2.String() == p1.Participant {
		p1, p2 = p2, p1
	} else {
		err = fmt.Errorf("channel participants not match ")
	}
	return
}

//UpdateChannelBalanceProof update balance proof
func UpdateChannelBalanceProof(participant, partner common.Address, lockedAmount *big.Int, partnerBalanceProof *BalanceProof) (err error) {
	c, err := getChannel(partnerBalanceProof.ChannelID.String())
	if err != nil {
		return
	}
	p1, p, err := verifyParticipants(c, participant, partner)
	if err != nil {
		return
	}
	if p.Nonce >= partnerBalanceProof.Nonce {
		return fmt.Errorf("nonce not match,now=%d,got=%d", p.Nonce, partnerBalanceProof.Nonce)
	}
	bi := stringToBigInt(p.TransferedAmount)
	if bi.Cmp(partnerBalanceProof.TransferAmount) > 0 {
		return fmt.Errorf("transfer amount cannot decrease now=%s,got=%s", bi, partnerBalanceProof.TransferAmount)
	}
	p.Nonce = partnerBalanceProof.Nonce
	p.TransferedAmount = bigIntToString(partnerBalanceProof.TransferAmount)
	p.LockedAmount = bigIntToString(lockedAmount)
	return updateBalance(p, p1)

}

//UpdateChannelDeposit 链上发生了deposit事件,需要更新信息
func UpdateChannelDeposit(channelIdentifier common.Hash, participant common.Address, deposit *big.Int) (err error) {
	c, err := getChannel(channelIdentifier.String())
	if err != nil {
		return
	}
	//相信来自链上的数据
	p := c.Participants[0]
	if p.Participant != participant.String() {
		p = c.Participants[1]
	}
	p.Deposit = bigIntToString(deposit)
	return updateBalance(c.Participants[0], c.Participants[1])
}

//CloseChannel because of channel closed event
func CloseChannel(channelIdentifier common.Hash) (err error) {
	c, err := getChannel(channelIdentifier.String())
	if err != nil {
		return err
	}
	c.Status = ChannelStatusClosed
	return db.Model(c).UpdateColumn("status", c.Status).Error
}

//SettleChannel because of channel settled event
func SettleChannel(channelIdentifier common.Hash) (err error) {
	c, err := getChannel(channelIdentifier.String())
	if err != nil {
		return err
	}
	c.Status = ChannelStatusSettled
	tx := db.Begin()
	err = tx.Delete(c).Error
	if err != nil {
		tx.Rollback()
		return
	}
	s := &SettledChannel{
		ChannelID:    c.ChannelID,
		Participant1: c.Participants[0].Participant,
		Participant2: c.Participants[1].Participant,
	}
	raw, err := json.Marshal(c)
	if err != nil {
		tx.Rollback()
		return
	}
	s.Data = string(raw)
	err = tx.Create(s).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return nil

}

//WithDrawChannel because of withdraw event
func WithDrawChannel(channelIdentifier common.Hash, p1Address, p2Address common.Address, p1Balance, p2Balance *big.Int) (err error) {
	c, err := getChannel(channelIdentifier.String())
	if err != nil {
		return err
	}
	c.Status = ChannelStatusOpen
	p1, p2 := c.Participants[0], c.Participants[1]
	//假定来自链上的数据不会造假
	if p1.Participant == p2Address.String() {
		p1, p2 = p2, p1
	}
	p1.Deposit = bigIntToString(p1Balance)
	p1.Nonce = 0
	p1.TransferedAmount = utils.BigInt0.String()
	p1.LockedAmount = utils.BigInt0.String()
	p1.Balance = p1.Deposit
	p2.Deposit = bigIntToString(p2Balance)
	p2.Nonce = 0
	p2.TransferedAmount = utils.BigInt0.String()
	p2.LockedAmount = utils.BigInt0.String()
	p2.Balance = bigIntToString(p2Balance)
	return db.Save(c).Error
}
func bigIntToString(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}
func stringToBigInt(s string) *big.Int {
	bi, b := new(big.Int).SetString(s, 10)
	if !b {
		bi = new(big.Int)
	}
	return bi
}

func updateBalance(p1, p2 *ChannelParticipantInfo) (err error) {
	p1TransferAmount := stringToBigInt(p1.TransferedAmount)
	p1LockedAmount := stringToBigInt(p1.LockedAmount)
	p1Deposit := stringToBigInt(p1.Deposit)
	p2TransferAmount := stringToBigInt(p2.TransferedAmount)
	p2LockedAmount := stringToBigInt(p2.LockedAmount)
	p2Deposit := stringToBigInt(p2.Deposit)
	p1Balance := p1Deposit.Add(p1Deposit, p2TransferAmount).Sub(p1Deposit, p1TransferAmount).Sub(p1Deposit, p1LockedAmount)
	p2Balance := p2Deposit.Add(p2Deposit, p1TransferAmount).Sub(p2Deposit, p2TransferAmount).Sub(p2Deposit, p2LockedAmount)
	p1.Balance = p1Balance.String()
	p2.Balance = p2Balance.String()
	tx := db.Begin()
	err = tx.Save(p1).Error
	if err != nil {
		return
	}
	err = tx.Save(p2).Error
	if err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit().Error
}
