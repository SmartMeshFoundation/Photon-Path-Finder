package storage

import (
	"context"
	"database/sql"
	"time"
)

//feeRateSchema create tb_fee_rate
const feeRateSchema = `
CREATE TABLE IF NOT EXISTS tb_fee_rate(
	id	BIGSERIAL NOT NULL PRIMARY KEY,
	channel_id TEXT NOT NULL,
	peer_address TEXT NOT NULL,
	fee_rate TEXT NOT NULL,
	effitime BIGINT NOT NULL
);
`

// insertFeeRateSQL sql for insert-FeeRate
const insertFeeRateSQL = "" +
	"INSERT INTO tb_fee_rate(channel_id,peer_address,fee_rate,effitime) VALUES(" +
	"$1,$2,$3,$4)"

// selectLatestFeeRateSQL sql for select-FeeRate
const selectLatestFeeRateSQL = "" +
	"SELECT fee_rate,effitime FROM tb_fee_rate WHERE channel_id=$1 and peer_address=$2 ORDER BY effitime DESC LIMIT 1"

// updateLatestFeeRateSQL sql for select-FeeRate
const updateLatestFeeRateSQL = "" +
	"UPDATE tb_fee_rate SET fee_rate=$1,effitime=$2 WHERE channel_id=$3 and peer_address=$4"

// balanceStatements interactive with db-operation
type feeRateStatements struct {
	insertFeeRateStmt *sql.Stmt
	selectFeeRateStmt *sql.Stmt
	updateLatestFeeRateStmt *sql.Stmt
}

// prepare prepare tb_balance
func (s *feeRateStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(feeRateSchema)
	if err != nil {
		return
	}
	if s.insertFeeRateStmt, err = db.Prepare(insertFeeRateSQL); err != nil {
		return
	}
	if s.selectFeeRateStmt, err = db.Prepare(selectLatestFeeRateSQL); err != nil {
		return
	}
	if s.updateLatestFeeRateStmt, err = db.Prepare(updateLatestFeeRateSQL); err != nil {
		return
	}
	return
}

// insertFeeRate insert data
func (s *feeRateStatements) insertFeeRate(ctx context.Context,
	channeID, peerAddress, feeRate string,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	stmt := s.updateLatestFeeRateStmt
	_, err = stmt.Exec(feeRate, timeMs,channeID, peerAddress)

	return
}

// selectLatestFeeRate select data
func (s *feeRateStatements) selectLatestFeeRate(ctx context.Context, channeID, peerAddress string) (
	feeRate string, effitime int64, err error) {
	stmt := s.selectFeeRateStmt
	err = stmt.QueryRow(channeID, peerAddress).Scan(&feeRate, &effitime)
	if err != nil {
		if err == sql.ErrNoRows {
			//log.WithError(err).Error("Unable to retrieve tLatestFeeRate from the db")
			return "0", 0, nil
		}
	}
	return
}
