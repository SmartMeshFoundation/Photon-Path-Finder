package rest

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/nkbai/goutils"

	"github.com/SmartMeshFoundation/Photon/log"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/ethereum/go-ethereum/common"
)

//balanceProofRequest is the json request for BalanceProof
type balanceProofRequest struct {
	BalanceProof     *model.BalanceProof `json:"balance_proof"`
	BalanceSignature []byte              `json:"balance_signature"`
	ProofSigner      common.Address      `json:"proof_signer"`
	LockedAmount     *big.Int            `json:"lock_amount"`
}

// UpdateBalanceProof handle the request with balance proof,implements GET and POST /balance
func UpdateBalanceProof(w rest.ResponseWriter, r *rest.Request) {
	var req = &balanceProofRequest{}
	var err error
	defer func() {
		log.Trace(fmt.Sprintf("UpdateBalanceProof op req=%s,err=%s", utils.StringInterface(req, 5), err))
	}()
	peer := r.PathParam("peer")
	peerAddress := common.HexToAddress(peer)
	err = r.DecodeJsonPayload(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//var locksAmount *big.Int
	partner, err := verifyBalanceProofSignature(req, peerAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.BalanceProof!=nil && req.BalanceProof.Nonce>0 {
		ce.HandleReceiveUserUpdateBalanceProof(peerAddress, partner, req.LockedAmount, req.BalanceProof)
	}
	err = w.WriteJson(nil)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}
