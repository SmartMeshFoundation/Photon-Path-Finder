package routing

import (
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/ethereum/go-ethereum/common"
)

// SetFeeRateRequest is the json request for SetFeeRate
type SetFeeRateRequest struct {
	ChannelID common.Hash `json:"channel_id"`
	FeeRate   string      `json:"fee_rate"`
	Signature []byte      `json:"signature"`
}

// GetFeeRateRequest is the json request for GetFeeRate
// Reponse all data of fee rate if channel_id is null
type GetFeeRateRequest struct {
	ObtainObj common.Address `json:"obtain_obj"`
	ChannelID []common.Hash  `json:"channel_id"`
}

// GetFeeRateRequest is the json request for GetFeeRate
type GetFeeRateResponse struct {
	Result map[common.Hash]*FeeRateInfo `json:"result"`
}

// FeeRateInfo stuct of fee-rate-info
type FeeRateInfo struct {
	ChannelID     common.Hash    `json:"channel_id"`
	PeerAddress   common.Address `json:"peer_address"`
	FeeRate       string         `json:"fee_rate"`
	EffectiveTime int64          `json:"effective_time"`
}


// SetFeeRate save request data of set_fee_rate
func SetFeeRate(req *http.Request,cfg config.PathFinder,feeRateDB *storage.Database,peerAddress string) util.JSONResponse {

	if req.Method != http.MethodPost || req.Method != http.MethodPut {
		return util.JSONResponse{
			Code: http.StatusMethodNotAllowed,
			JSON: util.NotFound("Bad method"),
		}
	}

	var r SetFeeRateRequest
	resErr := util.UnmarshalJSONRequest(req, &r)
	if resErr != nil {
		return *resErr
	}

	//validate json-input
	err := verifySinatureFeeRate(r, common.HexToAddress(peerAddress))
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: err.Error(), //util.BadJSON("peerAddress must be provided"),
		}
	}

	util.GetLogger(req.Context()).WithField("set_fee_rate", r.Signature).Info("Processing set_fee_rate request")

	var channelid = r.ChannelID
	var peeraddress = peerAddress
	var feerate= r.FeeRate
	err = feeRateDB.SaveRateFeeStorage(nil, channelid.String(), peeraddress, feerate)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: util.InvalidArgumentValue(err.Error()),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: util.OkJSON("true"),
	}

}

// GetFeeRate reponse fee_rate data
func GetFeeRate(req *http.Request,feeRateDB *storage.Database,peerAddress string) util.JSONResponse {
	if req.Method==http.MethodGet{
		if common.IsHexAddress(peerAddress){
			return util.JSONResponse{
				Code: http.StatusBadRequest,
				JSON: util.BadJSON("peerAddress must be provided"),
			}
		}

		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: util.InvalidArgumentValue("true"),
		}
	}

	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}