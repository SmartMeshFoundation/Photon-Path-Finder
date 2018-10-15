package storage

import (
	"database/sql"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"context"
	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
	common.PartitionOffsetStatements
	channelinfoStatement channelInfoStatements
}

// NewDatabase creates a new accounts and profiles database
func NewDatabase(dataSourceName string) (*Database, error) {
	var db *sql.DB
	var err error
	if db, err = sql.Open("postgres", dataSourceName); err != nil {
		return nil, err
	}
	partitions := common.PartitionOffsetStatements{}
	if err = partitions.Prepare(db, "pfs"); err != nil {
		return nil, err
	}

	cis := channelInfoStatements{}
	if err = cis.prepare(db); err != nil {
		return nil, err
	}
	return &Database{db, partitions, cis}, nil
}

//CreateChannel 通道创建时候建立通道双方的关系图
func (d *Database) CreateChannelInfoStorage(ctx context.Context,channelID,status,participant,partner string)  (err error){
	err=d.channelinfoStatement.createChannelInfo(ctx,status,channelID,participant,partner,0,0)
	return
}

//UpdateChannelStateStorage更新通道状态
func (d *Database) UpdateChannelStatusStorage(ctx context.Context,channelID,status string)  (err error){
	/*err=d.channelinfoStatement.updateChannelStatus(ctx,status,channelID)
	return*/
	return nil
}

//UpdateBalance 更新余额
func (d *Database) UpdateChannelInfoStorage(ctx context.Context,
	channelID,state,participant string,participant_capacity interface{})  (err error){
	/*err=d.channelinfoStatement.updateChannelDeposit(ctx,state,channelID,participant,participant_capacity)
	return*/
	return nil
}

//WithdrawChannelInfoStorage
func (d *Database) WithdrawChannelInfoStorage(ctx context.Context,channelID,status,participant string,capacity interface{},)  (err error){
	/*err=d.channelinfoStatement.updateChannelDeposit(ctx,status,channelID,participant,capacity)
	return*/
	return nil
}