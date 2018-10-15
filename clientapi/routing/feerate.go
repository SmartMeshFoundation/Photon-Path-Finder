package routing

import (
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
)

func SetFeeRate(req *http.Request,cfg config.PathFinder,feeRateDB *storage.Database,peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPost {

	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}

func GetFeeRate(req *http.Request,peerAddress string) util.JSONResponse {
	if req.Method==http.MethodGet{

	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}

}