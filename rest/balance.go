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
	// 如果节点启动参数中开启了ignore-mediatednode-request参数,那么该节点将不接收MediatedTransfer交易,此时需要报告给pfs,以免pfs把自己当中间节点来计算路由
	// 该参数没必要纳入签名
	IgnoreMediatedTransfer bool `json:"ignore_mediated_transfer"`
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
	if req.BalanceProof != nil && req.BalanceProof.Nonce > 0 {
		ce.HandleReceiveUserUpdateBalanceProof(peerAddress, partner, req.LockedAmount, req.BalanceProof, req.IgnoreMediatedTransfer)
	}
	err = w.WriteJson(nil)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}
