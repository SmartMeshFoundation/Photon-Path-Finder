package rest

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/ethereum/go-ethereum/common"
)

// pathRequest is the json request for GetPaths
type pathRequest struct {
	PeerFrom     common.Address `json:"peer_from"`
	PeerTo       common.Address `json:"peer_to"`
	TokenAddress common.Address `json:"token_address"`
	LimitPaths   int            `json:"limit_paths"`
	SendAmount   *big.Int       `json:"send_amount"`
	SortDemand   string         `json:"sort_demand"`
	Signature    []byte
}

// GetPaths handle the request with GetPaths,implements POST /paths
func GetPaths(w rest.ResponseWriter, r *rest.Request) {
	var req pathRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var peerFrom = req.PeerFrom
	var peerTo = req.PeerTo
	var tokenAddress = req.TokenAddress
	var limitPaths = req.LimitPaths
	var sendAmount = req.SendAmount
	var sortDemand = req.SortDemand
	pathResult, err := tn.GetPaths(peerFrom, peerTo, tokenAddress, sendAmount, limitPaths, sortDemand)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = w.WriteJson(pathResult)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}
