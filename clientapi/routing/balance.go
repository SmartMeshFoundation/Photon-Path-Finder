package routing

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/ethereum/go-ethereum/common"
)

// BalanceProof is the json request for BalanceProof
type BalanceProof struct {
	Nonce             int         `json:"nonce"`
	TransferredAmount *big.Int    `json:"transferred_amount"`
	ChannelID         common.Hash `json:"channel_id"`
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
	BalanceSignature []byte       `json:"balance_signature"`
	BalanceProof     BalanceProof `json:"balance_proof"`
	//Locks        []lock       `json:"locks"`
	LocksAmount *big.Int `json:"locks_amount"`
}

// UpdateBalanceProof handle the request with balance proof,implements GET and POST /balance
func UpdateBalanceProof(req *http.Request, ce blockchainlistener.ChainEvents, peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPut {
		var r balanceProofRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}
		//validate json-input
		if _, exist := ce.TokenNetwork.ChannelID2Address[r.BalanceProof.ChannelID]; !exist {
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: fmt.Sprintf("Unknown channel,channel_id=%s", r.BalanceProof.ChannelID.String()),
			}
		}

		var partner common.Address
		for _, xpartner := range ce.TokenNetwork.ChannelID2Address[r.BalanceProof.ChannelID] {
			if xpartner != common.HexToAddress(peerAddress) {
				partner = xpartner
				break
			}
		}

		//var locksAmount *big.Int
		err := verifySinature(&r, common.HexToAddress(peerAddress), partner)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusBadRequest,
				JSON: err.Error(),
			}
		}

		util.GetLogger(req.Context()).WithField("balance_proof", r.BalanceSignature).Info("Processing balance_proof request")

		err = ce.TokenNetwork.UpdateBalance(
			r.BalanceProof.ChannelID,
			partner,
			r.BalanceProof.Nonce,
			r.BalanceProof.TransferredAmount,
			r.LocksAmount)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: util.InvalidArgumentValue(err.Error()),
			}
		}
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: util.OkJSON("true"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}
