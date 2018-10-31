package routing

import (
	"net/http"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/util"
	"github.com/ethereum/go-ethereum/common"
)

// SetFeeRateRequest is the json request for SetFeeRate
type SetFeeRateRequest struct {
	ChannelID common.Hash `json:"channel_id"`
	FeeRate   int64       `json:"fee_rate"`
	Signature []byte      `json:"signature"`
}

// GetFeeRateRequest is the json request for GetFeeRate
// Reponse all data of fee rate if channel_id is null
type GetFeeRateRequest struct {
	ObtainObj common.Address `json:"obtain_obj"`
	ChannelID common.Hash    `json:"channel_id"`
	//Signature []byte         `json:"signature"` 查询完全没必要验证签名,这个信息应该是公开的,希望其他人知道的.
}

// GetFeeRateInfoResponse struct of fee-rate-info
type GetFeeRateInfoResponse struct {
	FeeRate int64 `json:"fee_rate"`
}

// SetFeeRate save request data of set_fee_rate
func SetFeeRate(req *http.Request, peerAddress string) util.JSONResponse {

	if req.Method != http.MethodPut {
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
	err := verifySinatureSetFeeRate(r, common.HexToAddress(peerAddress))
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: err.Error(), //util.BadJSON("peerAddress must be provided"),
		}
	}

	if r.FeeRate <= 0 {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: util.InvalidArgumentValue("fee_rate must be a  positive number"),
		}
	}

	util.GetLogger(req.Context()).WithField("set_fee_rate", r.Signature).Info("Processing set_fee_rate request")

	fee := &model.Fee{
		FeePolicy:  model.FeePolicyPercent,
		FeePercent: r.FeeRate,
	}
	err = model.UpdateChannelFeeRate(r.ChannelID, common.HexToAddress(peerAddress), fee)
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
func GetFeeRate(req *http.Request, peerAddress string, cfg config.PathFinder) util.JSONResponse {
	if req.Method == http.MethodPost {
		var r GetFeeRateRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}

		fee, err := model.GetChannelFeeRate(r.ChannelID, r.ObtainObj)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusNotFound,
				JSON: util.NotFound("any fee-rate data found"),
			}
		}

		reslut := &GetFeeRateInfoResponse{
			FeeRate: fee.FeePercent,
		}
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: reslut,
		}
	}

	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}
