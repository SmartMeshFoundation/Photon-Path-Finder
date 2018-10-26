package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// channelInfoSchema create tb_channel_info
const channelInfoSchema = `
CREATE TABLE IF NOT EXISTS tb_channel_info(
    id	BIGSERIAL NOT NULL PRIMARY KEY,
	ts  BIGINT NOT NULL,
	token TEXT NOT NULL,
	channel_id TEXT NOT NULL,
	channel_status TEXT NOT NULL,
	participant1 TEXT NOT NULL,
	participant2 TEXT NOT NULL,
	p1_status TEXT NOT NULL,
	p1_transferamount BIGINT NOT NULL,
	p1_nonce BIGINT NOT NULL DEFAULT 0,
	p1_lockedamount BIGINT NOT NULL,
	p1_deposit BIGINT NOT NULL,
	p1_balance BIGINT NOT NULL,
	p2_status TEXT NOT NULL,
	p2_transferamount BIGINT NOT NULL,
	p2_nonce BIGINT NOT NULL DEFAULT 0,
	p2_lockedamount BIGINT NOT NULL,
	p2_deposit BIGINT NOT NULL,
	p2_balance BIGINT NOT NULL
);
`

/*DROP TRIGGER IF EXISTS "trigger_tb_channel_info_update" on tb_channel_info;
CREATE TRIGGER trigger_tb_channel_info_update AFTER UPDATE
ON tb_channel_info FOR EACH ROW EXECUTE PROCEDURE trigger_update_balance_sync ();

CREATE OR REPLACE FUNCTION trigger_update_balance_sync () RETURNS TRIGGER AS $$
DECLARE passed BOOLEAN ;
BEGIN
IF (TG_OP = 'UPDATE') THEN
IF NEW.p1_deposit!= OLD.p1_deposit THEN
UPDATE tb_channel_info SET p1_balance=NEW.p1_deposit-NEW.p1_transferamount+NEW.p2_transferamount-NEW.p1_lockedamount
END IF;
IF NEW.p1_transferamount!= OLD.p1_transferamount THEN
UPDATE tb_channel_info SET p1_balance=NEW.p1_deposit-NEW.p1_transferamount+NEW.p2_transferamount-NEW.p1_lockedamount
END IF;
END IF ;
RETURN NULL ;
END ; $$ LANGUAGE plpgsql VOLATILE COST 200;*/

const (
	// createChannelInfoSQL sql for create-ChannelInfo
	createChannelInfoSQL = "" +
		"INSERT INTO tb_channel_info(ts,token,channel_id,channel_status,participant1,participant2," +
		"p1_status,p1_transferamount,p1_nonce,p1_lockedamount,p1_deposit,p1_balance," +
		"p2_status,p2_transferamount,p2_nonce,p2_lockedamount,p2_deposit,p2_balance " +
		") VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)"

	// selectAllChannelInfoSQL sql for select-AllChannelInfo
	selectAllChannelInfoSQL = "" +
		"SELECT token,channel_id,channel_status,participant1,participant2," +
		"p1_status,p1_transferamount,p1_nonce,p1_lockedamount,p1_deposit,p1_balance," +
		"p2_status,p2_transferamount,p2_nonce,p2_lockedamount,p2_deposit,p2_balance " +
		"FROM tb_channel_info ORDER BY ts DESC"

	// selectChanneCountByChannelIDSQL sql for select-ChanneCount-ByChannelID
	selectTokenByChannelIDSQL = "" +
		"SELECT token FROM tb_channel_info WHERE " +
		"channel_id = $1 "

	// selectNonceByChannelID1SQL sql for select-nonce-ByChannelID
	selectNonceByChannelID1SQL = "" +
		"SELECT p1_nonce FROM tb_channel_info WHERE " +
		"channel_id = $1 AND participant1=$2 "

	// selectNonceByChannelID1SQL sql for select-nonce-ByChannelID
	selectNonceByChannelID2SQL = "" +
		"SELECT p2_nonce FROM tb_channel_info WHERE " +
		"channel_id = $1 AND participant2=$2 "

	// updateChannelStatusSQL sql for update-hannelStatus
	updateChannelStatusSQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 WHERE channel_id = $3 "

	// updateChannelDeposit1SQL sql for update-ChannelDeposit
	updateChannelDeposit1SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p1_deposit= $3 ,p1_balance = " +
		"$3 - p1_transferamount +p2_transferamount -p1_lockedamount " +
		"WHERE channel_id = $4 AND participant1=$5 "

	// updateChannelDeposit1SQL sql for update-ChannelDeposit
	updateChannelDeposit2SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p2_deposit= $3 ,p2_balance = " +
		"$3 - p2_transferamount +p1_transferamount -p2_lockedamount " +
		"WHERE channel_id = $4 AND participant2=$5 "

	// updateChannelDepositSQL sql for update-ChannelDeposit
	updateBalanceProof1SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p1_transferamount=$3,p1_nonce=$4,p1_lockedamount=$5," +
		"p1_balance = p1_deposit-$3+p2_transferamount-$5, p2_balance=p2_deposit-p2_transferamount+$3-p2_lockedamount " +
		"WHERE channel_id = $6 AND participant1=$7 "

	// updateChannelDepositSQL sql for update-ChannelDeposit
	updateBalanceProof2SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p2_transferamount=$3,p2_nonce=$4,p2_lockedamount=$5," +
		"p2_balance = p2_deposit-$3+p1_transferamount-$5, p1_balance=p1_deposit-p1_transferamount+$3-p1_lockedamount " +
		"WHERE channel_id = $6 AND participant2=$7 "

	// updateChannelWithdrawSQL sql for update-ChannelWithdraw
	updateChannelWithdraw1SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p1_deposit= $3,p1_balance= $3," +
		"p1_nonce=0,p1_transferamount=0,p1_lockedamount=0," +
		"p2_nonce=0,p2_transferamount=0,p2_lockedamount=0 " +
		"WHERE channel_id = $4 AND participant1=$5 "

	// updateChannelWithdrawSQL sql for update-ChannelWithdraw
	updateChannelWithdraw2SQL = "" +
		"UPDATE tb_channel_info SET ts = $1 ,channel_status = $2 ,p2_deposit= $3,p2_balance= $3," +
		"p2_nonce=0,p2_transferamount=0,p2_lockedamount=0," +
		"p1_nonce=0,p1_transferamount=0,p1_lockedamount=0 " +
		"WHERE channel_id = $4 AND participant2=$5 "

	// updateChannelWithdrawSQL sql for update-selectFeeJudgeSQL
	/*selectFeeJudgeSQL = "" +
	"SELECT tb_fee_rate.peer_address,tb_fee_rate.fee_rate," +
	"tb_channel_info.channel_id,tb_channel_info.channel_status," +
	"tb_channel_info.participant1,tb_channel_info.participant2,tb_channel_info.p1_balance,tb_channel_info.p2_balance " +
	"FROM tb_fee_rate LEFT OUTER JOIN tb_channel_info ON " +
	"tb_fee_rate.channel_id=tb_channel_info.channel_id WHERE tb_fee_rate.effitime IN (SELECT " +
	"MAX(effitime) FROM tb_fee_rate GROUP BY peer_address)"*/
	selectFeeJudgeSQL = "" +
		"SELECT tb_fee_rate.peer_address,tb_fee_rate.fee_rate," +
		"tb_channel_info.channel_id,tb_channel_info.channel_status," +
		"tb_channel_info.participant1,tb_channel_info.participant2,tb_channel_info.p1_balance,tb_channel_info.p2_balance " +
		"FROM tb_fee_rate LEFT OUTER JOIN tb_channel_info ON " +
		"tb_fee_rate.channel_id=tb_channel_info.channel_id WHERE tb_channel_info.token=$1"

	// initFeeTableSQL sql for insert-initFeeTableSQL
	initFeeTableSQL = "" +
		"INSERT INTO tb_fee_rate(channel_id,peer_address,fee_rate,effitime) VALUES(" +
		"$1,$2,$3,$4)"
)

// balanceStatements interactive with db-operation
type channelInfoStatements struct {
	createChannelInfoStmt       *sql.Stmt
	updateChannelStatusStmt     *sql.Stmt
	updateChannelDeposit1Stmt   *sql.Stmt
	updateChannelDeposit2Stmt   *sql.Stmt
	selectTokenByChannelIDStmt  *sql.Stmt
	selectAllChannelInfoStmt    *sql.Stmt
	updateBalanceProof1Stmt     *sql.Stmt
	updateBalanceProof2Stmt     *sql.Stmt
	updateChannelWithdraw1Stmt  *sql.Stmt
	updateChannelWithdraw2Stmt  *sql.Stmt
	selectNonceByChannelID1Stmt *sql.Stmt
	selectNonceByChannelID2Stmt *sql.Stmt
	selectFeeJudgeStmt          *sql.Stmt
	initFeeTableStmt            *sql.Stmt
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
	if s.selectTokenByChannelIDStmt, err = db.Prepare(selectTokenByChannelIDSQL); err != nil {
		return
	}
	if s.updateChannelStatusStmt, err = db.Prepare(updateChannelStatusSQL); err != nil {
		return
	}
	if s.updateChannelDeposit1Stmt, err = db.Prepare(updateChannelDeposit1SQL); err != nil {
		return
	}
	if s.updateChannelDeposit2Stmt, err = db.Prepare(updateChannelDeposit2SQL); err != nil {
		return
	}
	if s.selectAllChannelInfoStmt, err = db.Prepare(selectAllChannelInfoSQL); err != nil {
		return
	}
	if s.updateBalanceProof1Stmt, err = db.Prepare(updateBalanceProof1SQL); err != nil {
		return
	}
	if s.updateBalanceProof2Stmt, err = db.Prepare(updateBalanceProof2SQL); err != nil {
		return
	}
	if s.updateChannelWithdraw1Stmt, err = db.Prepare(updateChannelWithdraw1SQL); err != nil {
		return
	}
	if s.updateChannelWithdraw2Stmt, err = db.Prepare(updateChannelWithdraw2SQL); err != nil {
		return
	}
	if s.selectNonceByChannelID1Stmt, err = db.Prepare(selectNonceByChannelID1SQL); err != nil {
		return
	}
	if s.selectNonceByChannelID2Stmt, err = db.Prepare(selectNonceByChannelID2SQL); err != nil {
		return
	}
	if s.selectFeeJudgeStmt, err = db.Prepare(selectFeeJudgeSQL); err != nil {
		return
	}
	if s.initFeeTableStmt, err = db.Prepare(initFeeTableSQL); err != nil {
		return
	}

	return
}

// selectTokenByChannelID select data
func (s *channelInfoStatements) selectTokenByChannelID(ctx context.Context, channeID string) (
	token string, err error) {
	stmt := s.selectTokenByChannelIDStmt
	err = stmt.QueryRow(channeID).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			//log.WithError(err).Error("Unable to retrieve token by channel_id from the db")
			return "", nil
		}
	}
	return
}

// selectOldNonceByChannelID get peer's nonce in some channel
func (s *channelInfoStatements) selectOldNonceByChannelID(ctx context.Context, channeID, peerAddress string, pIndex int) (
	nonce uint64, err error) {
	var stmt *sql.Stmt
	if pIndex == 1 {
		stmt = s.selectNonceByChannelID1Stmt
	} else if pIndex == 2 {
		stmt = s.selectNonceByChannelID2Stmt
	} else {
		err = fmt.Errorf("An error occurred in querying nonce, channel_id=%s ,peer_address=%s", channeID, peerAddress)
		return
	}
	err = stmt.QueryRow(channeID, peerAddress).Scan(&nonce)
	if err != nil {
		if err == sql.ErrNoRows {
			//log.WithError(err).Error("Unable to retrieve last nonce from the db")
			return 0, nil
		}
	}
	return
}

// initChannelInfo init db
func (s *channelInfoStatements) initChannelInfo(ctx context.Context, token, channelID, channelStatus, partipant1, partipant2 string,
	p1Status string, p1Transferamount int64, p1Nonce int, p1Lockedamount, p1Deposit, p1Balance int64,
	p2Status string, p2Transferamount int64, p2Nonce int, p2Lockedamount, p2Deposit, p2Balance int64,
	defaultFeeRate string,
) (err error) {
	xtoken, err := s.selectTokenByChannelID(nil, channelID)
	if err != nil {
		return
	}
	if xtoken == "" {
		stmt := s.createChannelInfoStmt
		timeMs := time.Now().UnixNano() / 1000000
		_, err = stmt.Exec(timeMs, token, channelID, channelStatus, partipant1, partipant2,
			p1Status, p1Transferamount, p1Nonce, p1Lockedamount, p1Deposit, p1Balance,
			p2Status, p2Transferamount, p2Nonce, p2Lockedamount, p2Deposit, p2Balance)
		stmtinitfee := s.initFeeTableStmt
		_, err = stmtinitfee.Exec(channelID, partipant1, defaultFeeRate, timeMs)
		_, err = stmtinitfee.Exec(channelID, partipant2, defaultFeeRate, timeMs)
	}
	return
}

// updateChannelStatus update data
func (s *channelInfoStatements) updateChannelStatus(ctx context.Context, status, channeID string,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	stmt := s.updateChannelStatusStmt
	_, err = stmt.Exec(timeMs, status, channeID)
	return
}

// updateChannelDeposit update data
func (s *channelInfoStatements) updateChannelDeposit(ctx context.Context,
	channeID, status, participant string, participantDeposit int64, pIndex int,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	var stmt *sql.Stmt
	if pIndex == 1 {
		stmt = s.updateChannelDeposit1Stmt
	} else if pIndex == 2 {
		stmt = s.updateChannelDeposit2Stmt
	} else {
		err = fmt.Errorf("An error occurred in ChannelDeposit, depositer=%s ,channel_id=%s", participant, channeID)
		return
	}
	_, err = stmt.Exec(
		timeMs, status, participantDeposit,
		channeID, participant,
	)
	return
}

// updateChannelWithdraw update data
func (s *channelInfoStatements) updateChannelWithdraw(ctx context.Context,
	channeID, status, participant string, participantWithdraw int64, pIndex int,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000
	var stmt *sql.Stmt
	if pIndex == 1 {
		stmt = s.updateChannelWithdraw1Stmt
	} else if pIndex == 2 {
		stmt = s.updateChannelWithdraw2Stmt
	} else {
		err = fmt.Errorf("An error occurred in ChannelWithdraw, Withdrawer=%s ,channel_id=%s", participant, channeID)
		return
	}
	_, err = stmt.Exec(
		timeMs, status, participantWithdraw,
		channeID, participant,
	)
	return
}

// updateBalanceProof update balance
func (s *channelInfoStatements) updateBalanceProof(ctx context.Context,
	channeID, status, participant string, transferredAmount, receivedAmount, lockedAmount int64, participantNonce uint64, pIndex int,
) (err error) {
	timeMs := time.Now().UnixNano() / 1000000

	var stmt *sql.Stmt
	if pIndex == 1 {
		stmt = s.updateBalanceProof1Stmt
	} else if pIndex == 2 {
		stmt = s.updateBalanceProof2Stmt
	} else {
		err = fmt.Errorf("An error occurred in UpdateBalanceProof, balanceProof object=%s ,channel_id=%s", participant, channeID)
		return
	}
	_, err = stmt.Exec(
		timeMs, status, transferredAmount, participantNonce, lockedAmount,
		channeID, participant,
	)
	return
}

// selectAllChannelInfo select data
func (s *channelInfoStatements) selectAllChannelInfo(ctx context.Context) (channelInfos []ChannelInfo, err error) {
	stmt := s.selectAllChannelInfoStmt
	rows, err := stmt.Query()
	if err != nil {
		return
	}
	channelInfos = []ChannelInfo{}

	defer rows.Close()
	for rows.Next() {
		var c ChannelInfo
		err = rows.Scan(
			&c.Token,
			&c.ChannelID,
			&c.ChannelStatus,
			&c.Partipant1,
			&c.Partipant2,
			&c.P1Status,
			&c.P1Transferamount,
			&c.P1Nonce,
			&c.P1Lockedamount,
			&c.P1Deposit,
			&c.P1Balance,
			&c.P2Status,
			&c.P2Transferamount,
			&c.P2Nonce,
			&c.P2Lockedamount,
			&c.P2Deposit,
			&c.P2Balance,
		)
		if err != nil {
			return
		}
		channelInfos = append(channelInfos, c)
	}
	return
}

// selectLatestJudgement select data
func (s *channelInfoStatements) selectLatestFeeJudge(ctx context.Context, tokenAddress string) (peerFeeAndBalances []*PeerFeeAndBalance, err error) {
	stmt := s.selectFeeJudgeStmt
	rows, err := stmt.Query(tokenAddress)
	if err != nil {
		return
	}
	peerFeeAndBalances = []*PeerFeeAndBalance{}
	defer rows.Close()
	for rows.Next() {
		var p PeerFeeAndBalance
		err = rows.Scan(
			&p.PeerAddr,
			&p.FeeRate,
			&p.ChannelID,
			&p.ChannelStatus,
			&p.Participant1,
			&p.Participant2,
			&p.P1Balance,
			&p.P2Balance,
		)
		if err != nil {
			continue
		}
		peerFeeAndBalances = append(peerFeeAndBalances, &p)
	}
	return
}
