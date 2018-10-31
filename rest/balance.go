package rest

import (
	"math/big"
	"net/http"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/ethereum/go-ethereum/common"
)

//balanceProofRequest is the json request for BalanceProof
type balanceProofRequest struct {
	BalanceProof     *model.BalanceProof `json:"balance_proof"`
	BalanceSignature []byte              `json:"balance_signature"`
	LockedAmount     *big.Int            `json:"lock_amount"`
}

// UpdateBalanceProof handle the request with balance proof,implements GET and POST /balance
func UpdateBalanceProof(w rest.ResponseWriter, r *rest.Request) {
	peer := r.PathParam("peer")
	peerAddress := common.HexToAddress(peer)
	var req = &balanceProofRequest{}
	err := r.DecodeJsonPayload(req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}

	//var locksAmount *big.Int
	partner, err := verifyBalanceProofSignature(req, peerAddress)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}

	err = tn.UpdateBalance(peerAddress, partner, req.LockedAmount, req.BalanceProof)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: nil,
	})
}
