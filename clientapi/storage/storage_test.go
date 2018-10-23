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

//func (d *Database) InitChannelInfoStorage(ctx context.Context, token, channelID, partipant1, partipant2 string) (err error) {
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

//func (d *Database) UpdateChannelStatusStorage(ctx context.Context, token, channelID, channelStatus, participant, partner string) (err error) {
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
