package storage

import (
	"bytes"
	"database/sql"
	"fmt"
	"testing"

	xcommon "github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
)

func setupDb(t *testing.T) *Database {
	var dataSourceName = "postgres://pfs:123456@localhost/pfs_xxx?sslmode=disable"
	var db *sql.DB
	var err error
	if db, err = sql.Open("postgres", dataSourceName); err != nil {
		return nil
	}
	partitions := xcommon.PartitionOffsetStatements{}
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
	return &Database{db, partitions, lbs, cis, tss, frs}
}

func TestNewDatabase(t *testing.T) {
	dataSourceName := "postgres://pfs:123456@localhost/pfs_xxx?sslmode=disable"
	_, err := NewDatabase(dataSourceName)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_SaveTokensStorage(t *testing.T) {
	db := setupDb(t)
	token := utils.NewRandomAddress()
	tokennetwork := utils.NewRandomAddress()
	err := db.SaveTokensStorage(nil, token.String(), tokennetwork.String())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokennetwork)
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
	db := setupDb(t)
	tokenmap, err := db.GetAllTokensStorage(nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokenmap)
}

func TestDatabase_GetAllChannelHistoryStorage(t *testing.T) {
	db := setupDb(t)
	cis, err := db.GetAllChannelHistoryStorage(nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(cis)
}

func TestDatabase_SaveLatestBlockNumberStorage(t *testing.T) {
	db := setupDb(t)
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
	db := setupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	partipant1 := utils.NewRandomAddress().String()
	partipant2 := utils.NewRandomAddress().String()
	err := db.InitChannelInfoStorage(nil, token, channelID, partipant1, partipant2)
	if err != nil {
		t.Error(err)
	}
}

func TestDatabase_UpdateChannelStatusStorage(t *testing.T) {
	db := setupDb(t)
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
	db := setupDb(t)
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
	db := setupDb(t)
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
	participant0:="0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	partner:="0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	participant1:="0xcccccccccccccccccccccccccccccccccccccccc"
	int0:=DbFideldIndex(participant0,partner)
	int1:=DbFideldIndex(participant1,partner)
	t.Log("The index after sorting (participant)is ",int0)
	t.Log("The index after sorting (participant)is ",int1)
}


func TestDatabase_UpdateBalanceProofStorage(t *testing.T) {
	db := setupDb(t)
	token := utils.NewRandomAddress().String()
	channelID := utils.NewRandomHash().String()
	status := "updatebalance"
	participant := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	transferredAmount:=int64(22)
	receivedAmount:=int64(11)
	lockedAmount:=int64(11)
	participantNonce:=1
	err := db.UpdateBalanceProofStorage(nil, token, channelID, status, participant, partner, transferredAmount,receivedAmount,lockedAmount,participantNonce)
	if err != nil {
		t.Error(err)
	}
	cis,err:=db.channelinfoStatement.selectAllChannelInfo(nil)
	if err != nil {
		t.Error(err)
	}
	for _,v:=range cis{
		if v.ChannelID==channelID{
			t.Logf(fmt.Sprintf("balance proof message: %s", utils.StringInterface(v, 2)))
			break
		}
	}
}

func TestDatabase_GetTokenByChannelID(t *testing.T) {
	db := setupDb(t)
	channelID := utils.NewRandomHash().String()
	token ,err := db.GetTokenByChannelID(nil, channelID)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	channelID="0x0398beea63f098e2d3bb59884be79eda00cf042e39ad65e5c43a0a280f969f93"
	tokenx ,err := db.GetTokenByChannelID(nil, channelID)
	if err != nil {
		t.Error(err)
	}
	t.Log(tokenx)
}

func TestDatabase_GetLastNonceByChannel(t *testing.T) {
	db := setupDb(t)
	channelID := utils.NewRandomHash().String()
	peerAddress := utils.NewRandomAddress().String()
	partner := utils.NewRandomAddress().String()
	nonce,err:=db.GetLastNonceByChannel(nil,channelID,peerAddress,partner)
	if err != nil {
		t.Error(err)
	}
	t.Log(nonce)
}

//func (d *Database) SaveRateFeeStorage(ctx context.Context, channelID, peerAddress, feeRate string) (err error) {
func TestDatabase_SaveRateFeeStorage(t *testing.T) {
	db := setupDb(t)
	channelID := utils.NewRandomHash().String()
	peerAddress := utils.NewRandomAddress().String()
	feeRate := "xxx$%^&*(()(56236547"
	err:=db.SaveRateFeeStorage(nil,channelID,peerAddress,feeRate)
	if err != nil {
		t.Error(err)
	}
	//
}