package storage

import (
	"context"
	"database/sql"
)

// channelInfoSchema create tb_latest_block_number
const latestBlockNumberSchema = `
CREATE TABLE IF NOT EXISTS tb_latest_block_number(
	latest_block_number BIGINT NOT NULL
);
`

const (
	// insertLatestBlockNumberSQL sql
	insertLatestBlockNumberSQL = "" +
		"INSERT INTO tb_latest_block_number(latest_block_number) VALUES($1)"

	// updatLatestBlockNumberSQL sql
	updatLatestBlockNumberSQL = "" +
		"UPDATE tb_latest_block_number SET latest_block_number = $1"

	// selectLatestBlockNumberSQL sql
	selectLatestBlockNumberSQL = "" +
		"SELECT latest_block_number FROM tb_latest_block_number"
)

// latestBlockNumberStatements interactive with db-operation
type latestBlockNumberStatements struct {
	insertLatestBlockNumberStmt *sql.Stmt
	updatLatestBlockNumberStmt  *sql.Stmt
	selectLatestBlockNumberStmt *sql.Stmt
}

// prepare prepare tb_latest_block_number
func (s *latestBlockNumberStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(latestBlockNumberSchema)
	if err != nil {
		return
	}
	if s.insertLatestBlockNumberStmt, err = db.Prepare(insertLatestBlockNumberSQL); err != nil {
		return
	}
	if s.updatLatestBlockNumberStmt, err = db.Prepare(updatLatestBlockNumberSQL); err != nil {
		return
	}
	if s.selectLatestBlockNumberStmt, err = db.Prepare(selectLatestBlockNumberSQL); err != nil {
		return
	}
	return
}

// insertLatestBlockNumber insert data
func (s *latestBlockNumberStatements) insertLatestBlockNumber(ctx context.Context, latestBlockNum int64,
) (err error) {
	stmt := s.insertLatestBlockNumberStmt
	_, err = stmt.Exec(latestBlockNum)
	if err != nil {
		return err
	}
	return nil
}

// updatLatestBlockNumber update LatestBlockNumber
func (s *latestBlockNumberStatements) updatLatestBlockNumber(ctx context.Context,
	latestBlockNum int64,
) (err error) {
	stmt := s.updatLatestBlockNumberStmt
	_, err = stmt.Exec(latestBlockNum)
	return
}

// selectLatestBlockNumber select LatestBlockNumber
func (s *latestBlockNumberStatements) selectLatestBlockNumber(ctx context.Context,
) (latestBlockNum int64, err error) {
	stmt := s.selectLatestBlockNumberStmt
	err = stmt.QueryRow().Scan(&latestBlockNum)
	if err != nil {
		if err == sql.ErrNoRows {
			//log.WithError(err).Error("Unable to retrieve LatestBlockNumber from the db")
			return 0, nil
		}
	}
	return
}
