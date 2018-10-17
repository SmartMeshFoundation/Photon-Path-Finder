package routing

import (
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"math/big"
	"regexp"
	"github.com/nkbai/dijkstra"
)

// feeRateRequest is the json request for GetPaths
type feeRateRequest1 struct {
	PeerFrom   string   `json:"peer_from"`
	PeerTo     string   `json:"peer_to"`
	LimitPaths int      `json:"limit_paths"`
	SendAmount *big.Int `json:"send_amount"`
	SortDemand string   `json:"sort_demand"`
	Sinature   []byte   `json:"signature"`
}

// feeRateRequest is the json response for GetPaths
type feeRate struct {
	PathID  int      `json:"path_id"`
	PathHop int      `json:"path_hop"`
	fee     int64    `json:"fee"`
	Result  []string `json:"result"`
}

// Graph Draw and cache n*n matrix
type Graph struct {
	Graph dijkstra.Graph
	//BalanceProof balanceProof
}

var(
	validAddressRegex = regexp.MustCompile(`^@(0x[0-9a-f]{40})`)
)
// GetPaths handle the request with GetPaths,implements POST /paths
func GetPaths(req *http.Request,peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPost {
		if len(peerAddress)!=22{
			return util.JSONResponse{
				Code:http.StatusNoContent,
				JSON:"peer address was incorrect",
			}
		}
/*		var checkValid
		ismatch:=validAddressRegex.MatchString(peerAddress)
		if !ismatch{
			checkValid=false
		}

		var obtainObj
		obtainObj=common.HexToAddress(peerAddress)
		if obtainObj*/
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}