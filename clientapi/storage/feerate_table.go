package storage

import "database/sql"

const feeRateSchema  = `
CREATE TABLE IF NOT EXISTS tb_feerate(
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	channe_id TEST NOT NULL,
	peer_address TEST NOT NULL,
	fee_rate TEST NOT NULL,
	effitime BIGINT NOT NULL,
);
CREATE SCHEMA_NAME 
`
const insertFeeRateSQL=""+
	"INSERT INTO tb_feerate(channel_id,peer_address,fee_rate,effitime) VALUES(" +
	"$1,$2,$3,$4,$5)"

const selectFeeRateSQL=""+
	"SELECT TOP(1) FROM tb_feerate WHERE channel_id=$1 and peer_address=$2 ORDER BY effitime DESC"

//获取所有的表，包括新建的tb_balance_1,tb_balance_2...
const selectAllTableNameByServer="SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"

// balanceStatements interactive with db-operation
type feeRateStatements struct {
	insertFeeRateStmt             *sql.Stmt
	selectFeeRateByChannelIDAndAddrStmt  *sql.Stmt
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
	if s.selectFeeRateByChannelIDAndAddrStmt, err = db.Prepare(selectFeeRateSQL); err != nil {
		return
	}
	return
}