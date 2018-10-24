package routing

import (
	"math/big"
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/ethereum/go-ethereum/common"
)

// pathRequest is the json request for GetPaths
type pathRequest struct {
	PeerFrom   common.Address `json:"peer_from"`
	PeerTo     common.Address `json:"peer_to"`
	LimitPaths int            `json:"limit_paths"`
	SendAmount *big.Int       `json:"send_amount"`
	SortDemand string         `json:"sort_demand"`
	Sinature   []byte         `json:"signature"`
}

// GetPaths handle the request with GetPaths,implements POST /paths
func GetPaths(req *http.Request, ce blockchainlistener.ChainEvents, peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPost {
		var r pathRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}
		//verify caller's sinature
		err := verifySinaturePaths(r, common.HexToAddress(peerAddress))
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusBadRequest,
				JSON: err.Error(),
			}
		}
		var peerFrom = r.PeerFrom
		var peerTo = r.PeerTo
		var limitPaths = r.LimitPaths
		var sendAmount = r.SendAmount
		var sortDemand = r.SortDemand
		pathResult, err := ce.TokenNetwork.GetPaths(peerFrom, peerTo, sendAmount, limitPaths, sortDemand)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusExpectationFailed,
				JSON: err.Error(),
			}
		}
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: pathResult,
		}
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}
