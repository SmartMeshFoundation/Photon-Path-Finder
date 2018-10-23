package routing

import (
	"math/big"
	"net/http"
	"regexp"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
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

/*// pathResult is the json response for GetPaths
type pathResult struct {
	PathID  int      `json:"path_id"`
	PathHop int      `json:"path_hop"`
	fee     int64    `json:"fee"`
	Result  []string `json:"result"`
}*/

// Graph Draw and cache n*n matrix
type Graph struct {
	Graph dijkstra.Graph
	//BalanceProof balanceProof
}

var (
	validAddressRegex = regexp.MustCompile(`^@(0x[0-9a-f]{40})`)
)

// GetPaths handle the request with GetPaths,implements POST /paths
func GetPaths(req *http.Request, ce blockchainlistener.ChainEvents, peerAddress string) util.JSONResponse {
	//	vmux.Handle("/{peerAddress}/paths",
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
				Code: http.StatusInternalServerError,
				JSON: err.Error(),
			}
		}
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: pathResult, //util.OkJSON("true"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}
