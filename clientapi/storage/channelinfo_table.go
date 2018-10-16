package storage

import (
	"database/sql"
	"context"
	"time"
	log "github.com/sirupsen/logrus"
)

// channelInfoSchema create tb_channel_info
const channelInfoSchema  = `
CREATE TABLE IF NOT EXISTS tb_channel_info_0(
    id	BIGSERIAL NOT NULL PRIMARY KEY,
	channel_id TEXT NOT NULL,
	ts  BIGINT NOT NULL,
	status	TEXT NOT NULL,
	participant TEXT NOT NULL,
	partner TEXT NOT NULL,
	participant_capacity BIGINT NOT NULL,
	partner_capacity BIGINT NOT NULL
);
`

const(
	// createChannelInfoSQL sql for create-ChannelInfo
	createChannelInfoSQL = ""+
	"INSERT INTO tb_channel_info_0(channel_id,ts,status,participant,partner,participant_capacity,partner_capacity) VALUES(" +
	"$1,$2,$3,$4,$5,$6,$7) "

	// selectAllChannelInfoSQL sql for select-AllChannelInfo
	selectAllChannelInfoSQL=""+
		"SELECT channel_id,status,participant,partner,participant_capacity,partner_capacity FROM tb_channel_info_0 ORDER BY ts DESC"

	// selectChanneCountByChannelIDSQL sql for select-ChanneCount-ByChannelID
	selectChanneCountByChannelIDSQL=""+
		"SELECT COUNT(*) FROM tb_channel_info_0 WHERE " +
		"channel_id = $1 AND participant = $2 AND partner = $3"

	// updateChannelStatusSQL sql for update-hannelStatus
	updateChannelStatusSQL = ""+
		"UPDATE tb_channel_info_0 SET ts = $1 ,status = $2 WHERE channel_id = $3 "

	// updateChannelDepositSQL sql for update-ChannelDeposit
	updateChannelDepositSQL = ""+
		"UPDATE tb_channel_info_0 SET ts = $1 ,status = $2 ,participant_capacity=$3 WHERE channel_id = $4 AND participant=$5 "

 	/*updateChannelInfo  = ""+
	"UPDATE tb_channel_info SET ts = $1 ,status = $2 ,participant_capacity = $3,partner_capacity = $4 WHERE channel_id = $5 AND " +
	"participant = $6 AND partner = $7 "*/
)

// balanceStatements interactive with db-operation
type channelInfoStatements struct {
	createChannelInfoStmt             *sql.Stmt
	updateChannelStatusStmt           *sql.Stmt
	updateChannelDepositStmt          *sql.Stmt
	selectChannelCountByChannelIDStmt *sql.Stmt
	selectAllChannelInfoStmt          *sql.Stmt
}

// prepare prepare tb_balance
func (s *channelInfoStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(channelInfoSchema)

	if err != nil {
		return
	}
	if s.createChannelInfoStmt, err = db.Prepare(createChannelInfoSQL); err != nil {
		return
	}
	if s.selectChannelCountByChannelIDStmt, err = db.Prepare(selectChanneCountByChannelIDSQL); err != nil {
		return
	}
	if s.updateChannelStatusStmt, err = db.Prepare(updateChannelStatusSQL); err != nil {
		return
	}
	if s.updateChannelDepositStmt, err = db.Prepare(updateChannelDepositSQL); err != nil {
		return
	}
	if s.selectAllChannelInfoStmt, err = db.Prepare(selectAllChannelInfoSQL); err != nil {
		return
	}

	return
}

// selectChannelCountByChannelID select data
func (s *channelInfoStatements)selectChannelCountByChannelID(ctx context.Context,channe_id,participant,partner string) (
	channelcount int64,
	err error) {
	stmt := s.selectChannelCountByChannelIDStmt
	err = stmt.QueryRow(channe_id,participant,partner).Scan(&channelcount)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Unable to retrieve LatestBlockNumber from the db")
		}
	}
	return
}

// createChannelInfo insert data
func (s *channelInfoStatements)initChannelInfo(ctx context.Context,
	channe_id,status,participant,partner string,participant_capacity,partner_capacity int64,
	) (err error) {
	timeMs:=time.Now().UnixNano()/1000000

	count,err:=s.selectChannelCountByChannelID(nil,channe_id,participant,partner)
	if err!=nil{
		return
	}
	if count==0{
		stmt:=s.createChannelInfoStmt
		_,err=stmt.Exec(channe_id,timeMs,status,participant,partner,participant_capacity,partner_capacity)
	}
	return
}

// updateChannelStatus update data
func (s *channelInfoStatements)updateChannelStatus(ctx context.Context,
	status,channe_id string,
) (err error) {
	timeMs:=time.Now().UnixNano()/1000000
	stmt:=s.updateChannelStatusStmt
	_,err=stmt.Exec(timeMs,status,channe_id)
	return
}

// updateBalance update data
func (s *channelInfoStatements)updateChannelDeposit(ctx context.Context,
	channe_id,status,participant string,participant_capacity int64,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	stmt := s.updateChannelDepositStmt

	_, err = stmt.Exec(
		timeMs, status, participant_capacity,
		channe_id, participant,
	)
	return
}

// selectChannelCountByChannelID select data
func (s *channelInfoStatements)selectAllChannelInfo(ctx context.Context,) (ChannelInfos []ChannelInfo, err error) {
	stmt := s.selectAllChannelInfoStmt
	//var channelID string
	//var status string
	//var participant string
	//var partner string
	//var participantCapacity int64
	//var partnerCapacity int64
	rows,err:=stmt.Query()
	if err!=nil{
		return
	}
	ChannelInfos =[]ChannelInfo{}

	defer rows.Close()
	for rows.Next(){
		var c ChannelInfo
		err=rows.Scan(&c.ChannelID,&c.Status,&c.Participant,&c.Partner,&c.ParticipantCapacity,&c.PartnerCapacity)
		if err!=nil{
			return
		}
		ChannelInfos=append(ChannelInfos,c)
	}
	return
}

