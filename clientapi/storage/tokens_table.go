package storage

import (
	"database/sql"
	"context"
	"github.com/ethereum/go-ethereum/common"
)

// tokensSchema create tb_tokens
const tokensSchema  = `
CREATE TABLE IF NOT EXISTS tb_tokens(
	id	BIGSERIAL NOT NULL PRIMARY KEY,
	token TEXT NOT NULL,
	token_network TEXT NOT NULL
);
-- CREATE SEQUENCE IF NOT EXISTS seq_tokens START 1;
`
const(
	// insertTokensSQL sql
	insertTokensSQL = ""+
		"INSERT INTO tb_tokens(token,token_network) VALUES($1,$2)"

	// selectLatestBlockNumberSQL sql
	selectAllTokensSQL=""+
		"SELECT token,token_network FROM tb_tokens"
	// selectCountTokenSQL
	selectCountTokenSQL=""+
		"SELECT COUNT(*) FROM tb_tokens where token=$1"

)
// latestBlockNumberStatements interactive with db-operation
type tokensStatements struct {
	insertTokensStmt     *sql.Stmt
	selectAllTokensStmt  *sql.Stmt
	selectCountTokenStmt *sql.Stmt
}

// prepare prepare tb_latest_block_number
func (s *tokensStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(tokensSchema)
	if err != nil {
		return
	}
	if s.insertTokensStmt, err = db.Prepare(insertTokensSQL); err != nil {
		return
	}
	if s.selectAllTokensStmt, err = db.Prepare(selectAllTokensSQL); err != nil {
		return
	}
	if s.selectCountTokenStmt, err = db.Prepare(selectCountTokenSQL); err != nil {
		return
	}
	return
}

// insertTokens save token2tokennetwork
func (s *tokensStatements)insertTokens(ctx context.Context,token,tokenNetwork string,
) (err error) {
	stmt0:=s.selectCountTokenStmt
	var tokencount int
	err = stmt0.QueryRow(token).Scan(&tokencount)
	if err!=nil{
		return err
	}
	if tokencount==0 {
		stmt := s.insertTokensStmt
		_, err = stmt.Exec(token, tokenNetwork)
		if err != nil {
			return err
		}
	}
	TokenNetwork2TokenMap[common.HexToAddress(tokenNetwork)]=common.HexToAddress(token)
	return nil
}

// selectTokens select data
func (s *tokensStatements)selectTokens(ctx context.Context) (addressmap AddressMap,err error) {
	stmt := s.selectAllTokensStmt
	rows, err := stmt.Query()
	if err != nil {
		return
	}
	addressmap = make(AddressMap)
	defer rows.Close()

	for rows.Next() {
		var token, tokenNetwork string
		err = rows.Scan(&token, &tokenNetwork)
		if err != nil {
			return
		}
		TokenNetwork2TokenMap[common.HexToAddress(tokenNetwork)]=common.HexToAddress(token)
		addressmap[common.HexToAddress(token)] = common.HexToAddress(tokenNetwork)
	}
	return
}




