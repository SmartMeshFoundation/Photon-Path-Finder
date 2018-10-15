package storage

import (
	"database/sql"
	"context"
	"time"
)

//channelInfoSchema
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
	createChannelInfoSQL = ""+
	"INSERT INTO tb_channel_info_0(channel_id,ts,status,participant,partner,participant_capacity,partner_capacity) VALUES(" +
	"$1,$2,$3,$4,$5,$6,$7) "

	/*updateChannelStatusSQL = ""+
		"UPDATE tb_channel_info SET ts = $1 ,status = $2 WHERE channel_id = $3 "

	updateChannelDepositSQL = ""+
		"UPDATE tb_channel_info SET ts = $1 ,status = $2 ,participant_capacity=$3 WHERE channel_id = $4 AND participant=$5 "

 	selectChannelInfoByChannelIDSQL=""+
	"SELECT status,participant,partner,participant_capacity,partner_capacity FROM tb_channel_info WHERE " +
	"channel_id = $1 "

 	updateChannelInfo  = ""+
	"UPDATE tb_channel_info SET ts = $1 ,status = $2 ,participant_capacity = $3,partner_capacity = $4 WHERE channel_id = $5 AND " +
	"participant = $6 AND partner = $7 "*/
)
// balanceStatements interactive with db-operation
type channelInfoStatements struct {
	createChannelInfoStmt            *sql.Stmt
	updateChannelStatusStmt          *sql.Stmt
	updateChannelDepositStmt         *sql.Stmt
	selectChannelInfoByChannelIDStmt *sql.Stmt
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
	/*if s.updateChannelStatusStmt, err = db.Prepare(updateChannelStatusSQL); err != nil {
		return
	}
	if s.updateChannelDepositStmt, err = db.Prepare(updateChannelDepositSQL); err != nil {
		return
	}
	if s.selectChannelInfoByChannelIDStmt, err = db.Prepare(selectChannelInfoByChannelIDSQL); err != nil {
		return
	}*/
	return
}

//createChannelInfo insert data
func (s *channelInfoStatements)createChannelInfo(ctx context.Context,
	status,channe_id,participant,partner string,participant_capacity,partner_capacity int64,
	) (err error) {
	timeMs:=time.Now().UnixNano()/1000000
	//判断是否已经存在
	//...
	stmt:=s.createChannelInfoStmt
	_,err=stmt.ExecContext(ctx,timeMs,status,channe_id,participant,partner,participant_capacity,partner_capacity)
	return
}

//updateChannelStatus insert data
func (s *channelInfoStatements)updateChannelStatus(ctx context.Context,
	status,channe_id string,
) (err error) {
	timeMs:=time.Now().UnixNano()/1000000
	stmt:=s.updateChannelStatusStmt
	_,err=stmt.ExecContext(ctx,timeMs,status,channe_id)
	return
}

//updateBalance
func (s *channelInfoStatements)updateChannelDeposit(ctx context.Context,
	channe_id,status,participant string,participant_capacity interface{},
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	stmt := s.updateChannelDepositStmt

	_, err = stmt.ExecContext(
		ctx, timeMs, status,
		channe_id,
		participant,
		participant_capacity,
		)
	return
}