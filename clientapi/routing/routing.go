package routing

import (
	"github.com/gorilla/mux"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
)

//Setup registers HTTP handlers with the given ServeMux.
func Setup(
	apiMux *mux.Router,
	cfg config.PathFinder,
	pfsdb *storage.Database,
	ce blockchainlistener.ChainEvents,
) {
	// "/versions"
	apiMux.Handle("/pathfinder/versions",
		common.MakeExternalAPI("versions", func(req *http.Request) util.JSONResponse {
			return util.JSONResponse{
				Code: http.StatusOK,
				JSON: struct {
					Versions []string `json:"versions"`
				}{[]string{
					"v1",
				}},
			}
		}),
	).Methods(http.MethodGet, http.MethodOptions)

	vmux := apiMux.PathPrefix("/pathfinder").Subrouter()

	// "/balance"
	vmux.Handle("/{peerAddress}/balance",
		common.MakeExternalAPI("update_balance_proof", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return UpdateBalanceProof(req,ce, vars["peerAddress"])
		}),
	).Methods(http.MethodPut, http.MethodOptions)

	// "/fee_rate"
	vmux.Handle("/{peerAddress}/fee_rate",
		common.MakeExternalAPI("set_fee_rate", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return SetFeeRate(req,cfg,pfsdb, vars["peerAddress"])
		}),
	).Methods(http.MethodPut, http.MethodOptions)

	// "/fee_rate"
	vmux.Handle("/{peerAddress}/fee_rate",
		common.MakeExternalAPI("get_fee_rate", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return GetFeeRate(req,pfsdb, vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/paths"
	vmux.Handle("/{peerAddress}/paths",
		common.MakeExternalAPI("get_paths", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return GetPaths(req, ce,vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/calc_signature"
	vmux.Handle("/{peerAddress}/calc_signature_balance_proof",
		common.MakeExternalAPI("calc_signature_for_test", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return SignDataForBalanceProof(req,vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)
}