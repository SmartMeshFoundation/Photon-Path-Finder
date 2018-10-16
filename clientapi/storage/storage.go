package storage

import (
	"database/sql"
	xcommon "github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"context"
	_ "github.com/lib/pq"
	"github.com/ethereum/go-ethereum/common"
)
// Database Data base
type Database struct {
	db *sql.DB
	xcommon.PartitionOffsetStatements
	latestBlockNumberStatement latestBlockNumberStatements
	channelinfoStatement       channelInfoStatements
	tokensStatement            tokensStatements
}

// ChannelInfo db-type as channel info
type ChannelInfo struct {
	ChannelID           string
	Status              string
	Participant         string
	Partner             string
	ParticipantCapacity int64
	PartnerCapacity     int64
}

//AddressMap is token address to mananger address
type AddressMap map[common.Address]common.Address

// NewDatabase creates a new accounts and profiles database
func NewDatabase(dataSourceName string) (*Database, error) {
	var db *sql.DB
	var err error
	if db, err = sql.Open("postgres", dataSourceName); err != nil {
		return nil, err
	}
	partitions := xcommon.PartitionOffsetStatements{}
	if err = partitions.Prepare(db, "pfs"); err != nil {
		return nil, err
	}

	//latest-block-number-db
	lbs := latestBlockNumberStatements{}
	if err = lbs.prepare(db); err != nil {
		return nil, err
	}
	//channel-info-db
	cis := channelInfoStatements{}
	if err = cis.prepare(db); err != nil {
		return nil, err
	}
	//channel-info-db
	tss := tokensStatements{}
	if err = tss.prepare(db); err != nil {
		return nil, err
	}

	return &Database{db, partitions, lbs, cis,tss}, nil
}

// SaveLatestBlockNumberStorage Save Latest BlockNumber Storage
func (d *Database) SaveLatestBlockNumberStorage(ctx context.Context,lastestblocknum int64)  (err error){
	err=d.latestBlockNumberStatement.updatLatestBlockNumber(ctx,lastestblocknum)
	return
}

// SaveTokensStorage Save Latest Tokens Storage
func (d *Database) SaveTokensStorage(ctx context.Context,token,tokennetwork string)  (err error) {
	err = d.tokensStatement.insertTokens(ctx, token, tokennetwork)
	return
}

// GetAllTokensStorage Get All Tokens Storage
func (d *Database) GetAllTokensStorage(ctx context.Context)  (token2TokenNetwork AddressMap,err error) {
	token2TokenNetwork = make(map[common.Address]common.Address)
	token2TokenNetwork, err = d.tokensStatement.selectTokens(ctx)
	return
}

// GetAllChannelHistoryStorage Get All ChannelHistory Storage
func (d *Database) GetAllChannelHistoryStorage(ctx context.Context)  (ChannelInfos []ChannelInfo ,err error){
	 ChannelInfos,err=d.channelinfoStatement.selectAllChannelInfo(ctx)
	 return
}

// GetLatestBlockNumberStorage Get Latest BlockNumber Storage
func (d *Database) GetLatestBlockNumberStorage(ctx context.Context)  (lastestnum int64, err error) {
	lastestnum, err = d.latestBlockNumberStatement.selectLatestBlockNumber(ctx)
	if lastestnum == -1 {
		err = d.latestBlockNumberStatement.insertLatestBlockNumber(ctx, 0)
		lastestnum, err = d.latestBlockNumberStatement.selectLatestBlockNumber(ctx)
	}
	return
}

// InitChannelInfoStorage Init ChannelInfo Storage
func (d *Database) InitChannelInfoStorage(ctx context.Context,channelID,status,participant,partner string)  (err error){
	err=d.channelinfoStatement.initChannelInfo(nil,channelID,status,participant,partner,0,0,)
	return
}

// UpdateChannelStatusStorage Update ChannelStatus Storage
func (d *Database) UpdateChannelStatusStorage(ctx context.Context,channelID,status,participant,partner string)  (err error){
	err=d.InitChannelInfoStorage(ctx,channelID,status,participant,partner)
	if err!=nil{
		return
	}
	err=d.channelinfoStatement.updateChannelStatus(ctx,status,channelID)
	return
}

// UpdateChannelInfoStorage Update ChannelInfo Storage
func (d *Database) UpdateChannelInfoStorage(ctx context.Context,
	channelID,status,participant ,partner string,participantCapacity int64)  (err error) {
	err = d.InitChannelInfoStorage(ctx, channelID, status, participant, partner)
	if err != nil {
		return
	}
	err=d.channelinfoStatement.updateChannelDeposit(ctx,channelID,status,participant,participantCapacity)
	return
}

// WithdrawChannelInfoStorage Withdraw ChannelInfo Storage
func (d *Database) WithdrawChannelInfoStorage(ctx context.Context,channelID,status,participant string,capacity interface{},)  (err error){
	/*err=d.channelinfoStatement.updateChannelDeposit(ctx,status,channelID,participant,capacity)
	return*/
	return nil
}