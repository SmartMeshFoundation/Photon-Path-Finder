package model

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	// import _ "github.com/jinzhu/gorm/dialects/postgres"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
	// import _ "github.com/jinzhu/gorm/dialects/mssql"
)

const (
	//ChannelStatusOpen 通道状态打开
	ChannelStatusOpen = iota + 1
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
}

/*
ChannelParticipantFee 存储通道一方的收费信息
*/
type ChannelParticipantFee struct {
	ID              int
	ChannelID       string `gorm:"index"`
	Participant     string `gorm:"index"`
	Token           string
	FeePolicy       int
	FeeConstantPart string //固定部分是一个整数,比如一次收取1token
	FeePercentPart  int64  //0表示不收费,1000表示收费千分之一
}

//BalanceValue return this participant's available balance
func (c *ChannelParticipantInfo) BalanceValue() *big.Int {
	return stringToBigInt(c.Balance)
}

//Fee return this participant's charge fee
func (c *ChannelParticipantInfo) Fee(token common.Address) *Fee {
	return GetChannelFeeRate(common.HexToHash(c.ChannelID), common.HexToAddress(c.Participant), token)
}

//Channel Channel基本信息
type Channel struct {
	ChannelID       string `gorm:"primary_key"`
	Token           string `gorm:"index"`
	OpenBlockNumber int64
	Status          int
	Participants    []*ChannelParticipantInfo `gorm:"ForeignKey:ChannelID"`
}

func orderParticipants(p1, p2 *ChannelParticipantInfo) (rp1, rp2 *ChannelParticipantInfo) {
	if p1.Participant < p2.Participant {
		return p1, p2
	}
	return p2, p1
}

//GetChannel from db
func GetChannel(channelID string) (c *Channel, err error) {
	c = &Channel{
		ChannelID: channelID,
	}
	err = db.Where(c).Preload("Participants").Find(c).Error
	if err != nil {
		return
	}
	c.Participants[0], c.Participants[1] = orderParticipants(c.Participants[0], c.Participants[1])
	return
}

//GetAllTokenChannels get all channels of this `token`
func GetAllTokenChannels(token common.Address) (cs []*Channel, err error) {
	err = db.Where(&Channel{
		Token:  token.String(),
		Status: ChannelStatusOpen,
	}).Preload("Participants").Find(&cs).Error
	if err != nil {
		return
	}
	for _, c := range cs {
		c.Participants[0], c.Participants[1] = orderParticipants(c.Participants[0], c.Participants[1])
	}
	return
}

//AddChannel add channel to db, 必须将相应的participant 信息清空.
func AddChannel(token, participant1, participant2 common.Address, ChannelIdentifier common.Hash, blockNumber int64) (c *Channel, err error) {
	channelID := ChannelIdentifier.String()
	c, err = GetChannel(channelID)
	if err == nil {
		err = fmt.Errorf("channelId %s duplicate", channelID)
		return
	}
	c = &Channel{ChannelID: channelID}
	c.Token = token.String()
	c.Status = ChannelStatusOpen
	c.OpenBlockNumber = blockNumber
	p1 := &ChannelParticipantInfo{
		Participant: participant1.String(),
	}
	p2 := &ChannelParticipantInfo{
		Participant: participant2.String(),
	}
	p1, p2 = orderParticipants(p1, p2)
	c.Participants = []*ChannelParticipantInfo{p1, p2}
	err = db.Create(c).Error
	return
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
func UpdateChannelBalanceProof(participant, partner common.Address, lockedAmount *big.Int, partnerBalanceProof *BalanceProof) (c *Channel, err error) {
	c, err = GetChannel(partnerBalanceProof.ChannelID.String())
	if err != nil {
		return
	}
	if c.OpenBlockNumber != partnerBalanceProof.OpenBlockNumber {
		err = fmt.Errorf("receive UpdateChannelBalanceProof on channel=%s,but open block number not match ,database openblocknumber=%d,balanceproof=%d",
			c.ChannelID, c.OpenBlockNumber, partnerBalanceProof.OpenBlockNumber,
		)
		return
	}
	p1, p, err := verifyParticipants(c, participant, partner)
	if err != nil {
		return
	}
	//在测试链上的时候,启用debug mode,这样节点删除数据库也不影响,否则会导致删除之后交易不能提交.
	if !params.DebugMode {
		if p.Nonce > partnerBalanceProof.Nonce {
			err = fmt.Errorf("nonce not match,now=%d,got=%d", p.Nonce, partnerBalanceProof.Nonce)
			return
		}
		if p.Nonce == partnerBalanceProof.Nonce {
			log.Info(fmt.Sprintf("duplicate nonce update"))
			return
		}
		bi := stringToBigInt(p.TransferedAmount)
		if bi.Cmp(partnerBalanceProof.TransferAmount) > 0 {
			err = fmt.Errorf("transfer amount cannot decrease now=%s,got=%s", bi, partnerBalanceProof.TransferAmount)
			return
		}
	}
	p.Nonce = partnerBalanceProof.Nonce
	p.TransferedAmount = bigIntToString(partnerBalanceProof.TransferAmount)
	p.LockedAmount = bigIntToString(lockedAmount)
	err = updateBalance(p, p1)
	//没必要再来查一次了,c中的就已经是最新的了
	//c, err = GetChannel(partnerBalanceProof.ChannelID.String())
	return
}

//UpdateChannelDeposit 链上发生了deposit事件,需要更新信息
func UpdateChannelDeposit(channelIdentifier common.Hash, participant common.Address, deposit *big.Int) (c *Channel, err error) {
	c, err = GetChannel(channelIdentifier.String())
	if err != nil {
		return
	}
	//相信来自链上的数据
	p := c.Participants[0]
	if p.Participant != participant.String() {
		p = c.Participants[1]
	}
	p.Deposit = bigIntToString(deposit)
	err = updateBalance(c.Participants[0], c.Participants[1])
	return
}

//CloseChannel because of channel closed event
func CloseChannel(channelIdentifier common.Hash) (c *Channel, err error) {
	c, err = GetChannel(channelIdentifier.String())
	if err != nil {
		return
	}
	c.Status = ChannelStatusClosed
	err = db.Model(c).UpdateColumn("status", c.Status).Error
	return
}

//SettleChannel because of channel settled event
func SettleChannel(channelIdentifier common.Hash) (c *Channel, err error) {
	c, err = GetChannel(channelIdentifier.String())
	if err != nil {
		return
	}
	c.Status = ChannelStatusSettled
	tx := db.Begin()
	err = tx.Delete(c).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Delete(c.Participants[0]).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Delete(c.Participants[1]).Error
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
	err = tx.Commit().Error
	return

}

//WithDrawChannel because of withdraw event
func WithDrawChannel(channelIdentifier common.Hash, p1Address, p2Address common.Address, p1Balance, p2Balance *big.Int, blockNumber int64) (c *Channel, err error) {
	c, err = GetChannel(channelIdentifier.String())
	if err != nil {
		return
	}
	c.OpenBlockNumber = blockNumber
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
	err = db.Save(c).Error
	return
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
	/*
		虽然正常情况下是不会出现负数,但是问题就在于如果一方无网,一方有网,就会出现负数的情况. 如果无网一方不提交balance proof,
		那么会造成有网一方也不能提交. 这会造成不必要的麻烦.
	*/
	if p1Balance.Cmp(utils.BigInt0) < 0 {
		p1.Balance = utils.BigInt0.String()
		//return fmt.Errorf("p1 %s balance is negative  %s", p1.Participant, p1Balance)
	}
	if p2Balance.Cmp(utils.BigInt0) < 0 {
		p2.Balance = utils.BigInt0.String()
		//return fmt.Errorf("p2 %s balance is negative %s", p2.Participant, p2Balance)
	}
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

func getDirectChannelFee(channelIdentifier common.Hash, participant common.Address) (cf *ChannelParticipantFee, err error) {
	cf = &ChannelParticipantFee{
		ChannelID:   channelIdentifier.String(),
		Participant: participant.String(),
	}
	err = db.Where(cf).Find(cf).Error
	return
}

//UpdateChannelFeeRate update channel's fee rate
func UpdateChannelFeeRate(channelIdentifier common.Hash, participant, token common.Address, fee *Fee) (err error) {
	cf, err := getDirectChannelFee(channelIdentifier, participant)
	if err != nil {
		cf = &ChannelParticipantFee{
			ChannelID:   channelIdentifier.String(),
			Token:       token.String(),
			Participant: participant.String(),
		}
	}
	cf.FeePolicy = fee.FeePolicy
	cf.FeeConstantPart = bigIntToString(fee.FeeConstant)
	cf.FeePercentPart = fee.FeePercent

	err = db.Save(cf).Error
	return
}

//GetChannelFeeRate get channel's fee rate
func GetChannelFeeRate(channelIdentifier common.Hash, participant, token common.Address) (fee *Fee) {
	cf, err := getDirectChannelFee(channelIdentifier, participant)
	if err == nil {
		fee = &Fee{
			FeePolicy:   cf.FeePolicy,
			FeeConstant: stringToBigInt(cf.FeeConstantPart),
			FeePercent:  cf.FeePercentPart,
		}
		return
	}
	//从来没有针对通道设置过
	fee, err = GetAccountTokenFee(participant, token)
	if err == nil {
		return
	}
	fee = GetAccountFeePolicy(participant)
	return
}
