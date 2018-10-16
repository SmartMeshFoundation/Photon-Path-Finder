package routing

import (
	"net/http"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"math/big"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
)

//balanceProof is the json request for BalanceProof
type BalanceProof struct {
	Nonce             int64       `json:"nonce"`
	ChannelID         common.Hash `json:"channel_id"`
	TransferredAmount *big.Int    `json:"transferred_amount"`
	LocksRoot         common.Hash `json:"locksroot"`
	AdditionalHash    common.Hash `json:"additional_hash"`
	Signature         []byte      `json:"signature"`
}

//lock is the json request for BalanceProof
type lock struct {
	LockedAmount *big.Int    `json:"locked_amount"`
	Expriation   *big.Int    `json:"expiration"`
	SecretHash   common.Hash `json:"secret_hash"`
}

//balanceProofRequest is the json request for BalanceProof
type balanceProofRequest struct {
	BalanceHash  []byte       `json:"balance_hash"`
	BalanceProof BalanceProof `json:"balance_proof"`
	Locks        []lock       `json:"locks"`
}

// Balance handle the request with balance proof,implements GET and POST /balance
func UpdateBalanceProof(req *http.Request,cfg config.PathFinder,db *storage.Database,ce blockchainlistener.ChainEvents, peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPut {
		var r balanceProofRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}

		//validate json-input
		var partner common.Address
		var locksAmount *big.Int
		partner, locksAmount, err := verifySinature(r, common.HexToAddress(peerAddress))
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusBadRequest,
				JSON: util.BadJSON("peerAddress must be provided"),
			}
		}

		util.GetLogger(req.Context()).WithField("balance_proof", r.BalanceHash).Info("Processing balance_proof request")

		err = ce.TokenNetwork.UpdateBalance(
			r.BalanceProof.ChannelID,
			partner,
			r.BalanceProof.Nonce,
			r.BalanceProof.TransferredAmount,
			locksAmount)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: util.InvalidArgumentValue("argument was incorrect"),
			}
		}
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: util.InvalidArgumentValue("ok"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}

