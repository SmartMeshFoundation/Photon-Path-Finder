package storage

import (
	"context"
	"database/sql"
	"sort"
	"strconv"

	xcommon "github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq"
)

// Database Data base
type Database struct {
	db *sql.DB
	xcommon.PartitionOffsetStatements
	latestBlockNumberStatement latestBlockNumberStatements
	channelinfoStatement       channelInfoStatements
	tokensStatement            tokensStatements
	feereteStatement           feeRateStatements
	feeRateDefault             string
}

// ChannelInfo db-type as channel info
type ChannelInfo struct {
	Token            string
	ChannelID        string
	ChannelStatus    string
	Partipant1       string
	Partipant2       string
	P1Status         string
	P1Transferamount int64
	P1Nonce          int
	P1Lockedamount   int64
	P1Deposit        int64
	P1Balance        int64
	P2Status         string
	P2Transferamount int64
	P2Nonce          int
	P2Lockedamount   int64
	P2Deposit        int64
	P2Balance        int64
}

// PeerFeeAndBalance db-type as peer's fee and balance
type PeerFeeAndBalance struct {
	PeerAddr      string
	FeeRate       string
	ChannelID     string
	ChannelStatus string
	Participant1  string
	Participant2  string
	P1Balance     int64
	P2Balance     int64
}

//AddressMap is token address to mananger address
type AddressMap map[common.Address]common.Address

var TokenNetwork2TokenMap map[common.Address]common.Address

// NewDatabase creates a new accounts and profiles database
func NewDatabase(dataSourceName, feeRateDefault string) (*Database, error) {
	var db *sql.DB
	var err error
	if db, err = sql.Open("postgres", dataSourceName); err != nil {
		return nil, err
	}
	_, err = strconv.ParseFloat(feeRateDefault, 32)
	if err != nil {
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
	//fee-rate-db
	frs := feeRateStatements{}
	if err = frs.prepare(db); err != nil {
		return nil, err
	}
	//channel-info-db
	cis := channelInfoStatements{}
	if err = cis.prepare(db); err != nil {
		return nil, err
	}
	//token-db
	tss := tokensStatements{}
	if err = tss.prepare(db); err != nil {
		return nil, err
	}

	TokenNetwork2TokenMap = make(map[common.Address]common.Address)
	return &Database{db, partitions, lbs, cis, tss, frs, feeRateDefault}, nil
}

// SaveTokensStorage Save Latest Tokens Storage
func (d *Database) SaveTokensStorage(ctx context.Context, token, tokennetwork string) (err error) {
	err = d.tokensStatement.insertTokens(ctx, token, tokennetwork)
	return
}

// GetAllTokensStorage Get All Tokens Storage
func (d *Database) GetAllTokensStorage(ctx context.Context) (token2TokenNetwork AddressMap, err error) {
	token2TokenNetwork = make(map[common.Address]common.Address)
	token2TokenNetwork, err = d.tokensStatement.selectTokens(ctx)
	return
}

// GetAllChannelHistoryStorage Get All ChannelHistory Storage
func (d *Database) GetAllChannelHistoryStorage(ctx context.Context) (channelInfos []ChannelInfo, err error) {
	channelInfos, err = d.channelinfoStatement.selectAllChannelInfo(ctx)
	return
}

// SaveLatestBlockNumberStorage Save Latest BlockNumber Storage
func (d *Database) SaveLatestBlockNumberStorage(ctx context.Context, lastestblocknum int64) (err error) {
	err = d.latestBlockNumberStatement.updatLatestBlockNumber(ctx, lastestblocknum)
	return
}

// GetLatestBlockNumberStorage Get Latest BlockNumber Storage
func (d *Database) GetLatestBlockNumberStorage(ctx context.Context) (lastestnum int64, err error) {
	lastestnum, err = d.latestBlockNumberStatement.selectLatestBlockNumber(ctx)
	if lastestnum == 0 {
		err = d.latestBlockNumberStatement.insertLatestBlockNumber(ctx, 0)
		lastestnum, err = d.latestBlockNumberStatement.selectLatestBlockNumber(ctx)
	}
	return
}

// InitChannelInfoStorage Init ChannelInfo Storage
func (d *Database) InitChannelInfoStorage(ctx context.Context, token, channelID, partipant1, partipant2 string) (err error) {
	//default p1=0xaa p2=0xbb
	partipantPairs := []string{partipant1, partipant2}
	sort.Strings(partipantPairs)

	err = d.channelinfoStatement.initChannelInfo(nil, token, channelID, "createchannel", partipantPairs[0], partipantPairs[1],
		"online", 0, 0, 0, 0, 0,
		"online", 0, 0, 0, 0, 0, d.feeRateDefault)
	return
}

// UpdateChannelStatusStorage Update ChannelStatus Storage
func (d *Database) UpdateChannelStatusStorage(ctx context.Context, token, channelID, channelStatus, participant, partner string) (err error) {
	err = d.InitChannelInfoStorage(ctx, token, channelID, participant, partner)
	if err != nil {
		return
	}
	err = d.channelinfoStatement.updateChannelStatus(ctx, channelStatus, channelID)
	return
}

// UpdateChannelDepositStorage Update Deposit Storage
func (d *Database) UpdateChannelDepositStorage(ctx context.Context, token,
	channelID, status, participant, partner string, participantDeposit int64) (err error) {
	err = d.InitChannelInfoStorage(ctx, token, channelID, participant, partner)
	if err != nil {
		return
	}
	fieldIndex := DbFideldIndex(participant, partner)
	err = d.channelinfoStatement.updateChannelDeposit(ctx, channelID, status, participant, participantDeposit, fieldIndex)
	return
}

// WithdrawChannelInfoStorage Withdraw ChannelInfo Storage
func (d *Database) WithdrawChannelInfoStorage(ctx context.Context, token, channelID, status, participant, partner string, participantCapacity int64) (err error) {
	err = d.InitChannelInfoStorage(ctx, token, channelID, participant, partner)
	if err != nil {
		return
	}
	fieldIndex := DbFideldIndex(participant, partner)
	err = d.channelinfoStatement.updateChannelWithdraw(ctx, channelID, status, participant, participantCapacity, fieldIndex)
	return
}

// DbFideldIndex get index of partipantPairs
func DbFideldIndex(participant, partner string) int {
	var fieldIndex = 1
	partipantPairs := []string{participant, partner}
	sort.Strings(partipantPairs)
	if common.HexToAddress(participant) != common.HexToAddress(partipantPairs[0]) {
		fieldIndex = 2
	}
	return fieldIndex
}

// UpdateBalanceProofStorage Update balance proof Storage
func (d *Database) UpdateBalanceProofStorage(ctx context.Context, token,
	channelID, status, participant, partner string, transferredAmount, receivedAmount, lockedAmount int64, participantNonce uint64) (err error) {
	err = d.InitChannelInfoStorage(ctx, token, channelID, participant, partner)
	if err != nil {
		return
	}
	fieldIndex := DbFideldIndex(participant, partner)
	err = d.channelinfoStatement.updateBalanceProof(ctx, channelID, status, participant, transferredAmount, receivedAmount, lockedAmount, participantNonce, fieldIndex)
	return
}

// GetTokenByChannelID Get token by channelID
func (d *Database) GetTokenByChannelID(ctx context.Context, channelID string) (token string, err error) {
	token, err = d.channelinfoStatement.selectTokenByChannelID(ctx, channelID)
	return
}

// GetLastNonceByChannel Get LastNonce By Channel
func (d *Database) GetLastNonceByChannel(ctx context.Context, channelID, peerAddress, partner string) (nonce uint64, err error) {
	fieldIndex := DbFideldIndex(peerAddress, partner)
	nonce, err = d.channelinfoStatement.selectOldNonceByChannelID(ctx, channelID, peerAddress, fieldIndex)
	return
}

// SaveRateFeeStorage Save Rate Fee Storage
func (d *Database) SaveRateFeeStorage(ctx context.Context, channelID, peerAddress, feeRate string) (err error) {
	err = d.feereteStatement.insertFeeRate(ctx, channelID, peerAddress, feeRate)
	return
}

// GetLastestRateFeeStorage Save Rate Fee Storage
func (d *Database) GetLastestRateFeeStorage(ctx context.Context, channelID, peerAddress string) (feeRate string, effitime int64, err error) {
	feeRate, effitime, err = d.feereteStatement.selectLatestFeeRate(ctx, channelID, peerAddress)
	return
}

// GetLatestFeeJudge refush peer's balance and lasest fee
func (d *Database) GetLatestFeeJudge(ctx context.Context, tokenAddress string) (peerFeeAndBalances []*PeerFeeAndBalance, err error) {
	peerFeeAndBalances, err = d.channelinfoStatement.selectLatestFeeJudge(ctx, tokenAddress)
	return
}
