package storage

import (
	"testing"
	"database/sql"
	xcommon "github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaidenbai/utils"
	"fmt"
)

func setupDb(t *testing.T) (*Database) {
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
	TokenNetwork2TokenMap=make(map[common.Address]common.Address)
	if err!=nil{
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
	token:=utils.EmptyAddress.String()
	tokennetwork:=utils.EmptyAddress.String()
	err:=db.SaveTokensStorage(nil,token,tokennetwork)
	if err!=nil{
		t.Error(err)
	}
}

func TestDatabase_GetAllTokensStorage(t *testing.T) {
	db := setupDb(t)
	tokenmap,err:=db.GetAllTokensStorage(nil)
	if err!=nil{
		t.Error(err)
	}
	fmt.Println(tokenmap)
}