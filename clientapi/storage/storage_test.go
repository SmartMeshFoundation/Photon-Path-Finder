package storage

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	ycommon "github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
)

func SetupDb(t *testing.T) *Database {
	var dataSourceName = "postgres://pfs:123456@localhost/pfs_xxx?sslmode=disable"
	var db *sql.DB
	var err error
	if db, err = sql.Open("postgres", dataSourceName); err != nil {
		return nil
	}
	partitions := ycommon.PartitionOffsetStatements{}
	if err = partitions.Prepare(db, "pfs"); err != nil {
		return nil
	}
	lbs := latestBlockNumberStatements{}
	if err = lbs.prepare(db); err != nil {
		return nil
	}
	cis := channelInfoStatements{}
	if err = cis.prepare(db); err != nil {
		return nil
	}
	tss := tokensStatements{}
	if err = tss.prepare(db); err != nil {
		return nil
	}
	frs := feeRateStatements{}
	if err = frs.prepare(db); err != nil {
		return nil
	}
	TokenNetwork2TokenMap = make(map[common.Address]common.Address)
	if err != nil {
		t.Error(err)
	}
	defaultFeeRate := "0.00001"
	return &Database{db, partitions, lbs, cis, tss, frs, defaultFeeRate}
}

func TestNewDatabase(t *testing.T) {
	dataSourceName := "postgres://pfs:123456@localhost/pfs_xxx?sslmode=disable"
	defaultFeeRate := "0.00001"
	_, err := NewDatabase(dataSourceName, defaultFeeRate)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_SaveTokensStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress()
	tokennetwork := utils.NewRandomAddress()
	err := db.SaveTokensStorage(nil, token.String(), tokennetwork.String())
	if err != nil {
		t.Error(err)
	}
	t.Log(tokennetwork)
	tm, err := db.GetAllTokensStorage(nil)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(tm[token].Bytes(), tokennetwork.Bytes()) {
		t.Error("save failed")
	}
}

func TestDatabase_GetAllTokensStorage(t *testing.T) {
	db := SetupDb(t)
	tokenmap, err := db.GetAllTokensStorage(nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(tokenmap)
}

func TestDatabase_GetAllChannelHistoryStorage(t *testing.T) {
	db := SetupDb(t)
	cis, err := db.GetAllChannelHistoryStorage(nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(cis)
}

func TestDatabase_SaveLatestBlockNumberStorage(t *testing.T) {
	db := SetupDb(t)
	err := db.SaveLatestBlockNumberStorage(nil, 555)
	if err != nil {
		t.Error(err)
	}
	num, err := db.GetLatestBlockNumberStorage(nil)
	if err != nil {
		t.Error(err)
	}
	if num != 555 {
		t.Error("get LatestBlockNumber failed")
	}
}

func TestDatabase_InitChannelInfoStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	partipant1 := utils.NewRandomAddress().String()
	partipant2 := utils.NewRandomAddress().String()
	err := db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
	//path test data
	token = "0x1234567890123456789012345678901234567890"
	channelID = "0x1212121212121212121212121212121212121212121212121212121212121212"
	partipant1 = "0x1111111111111111111111111111111111111111"
	partipant2 = "0x2222222222222222222222222222222222222222"
	err = db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
	token = "0x1234567890123456789012345678901234567890"
	channelID = "0x2323232323232323232323232323232323232323232323232323232323232323"
	partipant1 = "0x2222222222222222222222222222222222222222"
	partipant2 = "0x3333333333333333333333333333333333333333"
	err = db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
	token = "0x1234567890123456789012345678901234567890"
	channelID = "0x3434343434343434343434343434343434343434343434343434343434343434"
	partipant1 = "0x3333333333333333333333333333333333333333"
	partipant2 = "0x4444444444444444444444444444444444444444"
	err = db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
	token = "0x1234567890123456789012345678901234567890"
	channelID = "0x4545454545454545454545454545454545454545454545454545454545454545"
	partipant1 = "0x4444444444444444444444444444444444444444"
	partipant2 = "0x5555555555555555555555555555555555555555"
	err = db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
	token = "0x1234567890123456789012345678901234567890"
	channelID = "0x5656565656565656565656565656565656565656565656565656565656565656"
	partipant1 = "0x5555555555555555555555555555555555555555"
	partipant2 = "0x6666666666666666666666666666666666666666"
	err = db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_UpdateChannelStatusStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	channelStatus := "close"
	participant := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	err := db.UpdateChannelStatusStorage(nil, token, channelID, channelStatus, participant, partner)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_UpdateChannelDepositStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	status := "deposit"
	participant := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	var participantDeposit int64
	err := db.UpdateChannelDepositStorage(nil, token, channelID, status, participant, partner, participantDeposit)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_WithdrawChannelInfoStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	status := "withdraw"
	participant := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	var participantDeposit int64
	err := db.WithdrawChannelInfoStorage(nil, token, channelID, status, participant, partner, participantDeposit)
	if err != nil {
		t.Error(err)
	}
}

func TestDbFideldIndex(t *testing.T) {
	participant0 := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	partner := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	participant1 := "0xcccccccccccccccccccccccccccccccccccccccc"
	int0 := DbFideldIndex(participant0, partner)
	int1 := DbFideldIndex(participant1, partner)
	t.Log("The index after sorting (participant)is ", int0)
	t.Log("The index after sorting (participant)is ", int1)
}

func TestDatabase_UpdateBalanceProofStorage(t *testing.T) {
	db := SetupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	status := "updatebalance"
	participant := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	transferredAmount := int64(22)
	receivedAmount := int64(11)
	lockedAmount := int64(11)
	participantNonce := uint64(5)
	err := db.UpdateBalanceProofStorage(nil, token, channelID, status, participant, partner, transferredAmount, receivedAmount, lockedAmount, participantNonce)
	if err != nil {
		t.Error(err)
	}
	cis, err := db.channelinfoStatement.selectAllChannelInfo(nil)
	if err != nil {
		t.Error(err)
	}
	for _, v := range cis {
		if v.ChannelID == channelID {
			t.Logf("balance proof message: %s", utils.StringInterface(v, 2))
			break
		}
	}
}

func TestDatabase_GetTokenByChannelID(t *testing.T) {
	db := SetupDb(t)
	channelID := utils.NewRandomHash().String()
	token, err := db.GetTokenByChannelID(nil, channelID)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	channelID = "0x0398beea63f098e2d3bb59884be79eda00cf042e39ad65e5c43a0a280f969f93"
	tokenx, err := db.GetTokenByChannelID(nil, channelID)
	if err != nil {
		t.Error(err)
	}
	t.Log(tokenx)
}

func TestDatabase_GetLastNonceByChannel(t *testing.T) {
	db := SetupDb(t)
	channelID := utils.NewRandomHash().String()
	peerAddress := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	nonce, err := db.GetLastNonceByChannel(nil, channelID, peerAddress, partner)
	if err != nil {
		t.Error(err)
	}
	t.Log(nonce)
}

func TestDatabase_SaveRateFeeStorage(t *testing.T) {
	db := SetupDb(t)
	channelID := utils.NewRandomHash().String()
	peerAddress := utils.NewRandomAddress().String()
	feeRate := "0.099"
	err := db.SaveRateFeeStorage(nil, channelID, peerAddress, feeRate)
	if err != nil {
		t.Error(err)
	}
	saveok := false
	// random(address,hash) may cause mismatch of tables:channel_info and fee_rate
	pfb, err := db.GetLatestFeeJudge(nil, "0x")
	if err != nil {
		if strings.Index(err.Error(), "*string") == -1 {
			t.Error(err)
		}
	}
	for _, v := range pfb {
		if v.PeerAddr == peerAddress {
			saveok = true
			break
		}
	}
	t.Log(saveok)
}
