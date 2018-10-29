package routing

import (
	"net/http"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/util"
	"github.com/gorilla/mux"
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
			return UpdateBalanceProof(req, ce, vars["peerAddress"])
		}),
	).Methods(http.MethodPut, http.MethodOptions)

	// "/fee_rate"
	vmux.Handle("/{peerAddress}/set_fee_rate",
		common.MakeExternalAPI("set_fee_rate", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return SetFeeRate(req, pfsdb, vars["peerAddress"])
		}),
	).Methods(http.MethodPut, http.MethodOptions)

	// "/fee_rate"
	vmux.Handle("/{peerAddress}/get_fee_rate",
		common.MakeExternalAPI("get_fee_rate", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return GetFeeRate(req, pfsdb, vars["peerAddress"], cfg)
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/paths"
	vmux.Handle("/{peerAddress}/paths",
		common.MakeExternalAPI("get_paths", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return GetPaths(req, ce, vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/calc_signature_set_fee"
	vmux.Handle("/{peerAddress}/calc_signature_set_fee",
		common.MakeExternalAPI("calc_signature_setfee_for_test", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return signDataForSetFee(req, cfg, vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/calc_signature_get_fee"
	vmux.Handle("/{peerAddress}/calc_signature_get_fee",
		common.MakeExternalAPI("calc_signature_getfee_for_test", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return signDataForGetFee(req, cfg, vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)

	// "/calc_signature_paths"
	vmux.Handle("/{peerAddress}/calc_signature_paths",
		common.MakeExternalAPI("calc_signature_paths_for_test", func(req *http.Request) util.JSONResponse {
			vars := mux.Vars(req)
			return signDataForPath(req, cfg, vars["peerAddress"])
		}),
	).Methods(http.MethodPost, http.MethodOptions)
}
