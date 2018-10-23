package routing

import (
	"net/http"
	"strconv"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
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
	ChannelID common.Hash    `json:"channel_id"`
	Signature []byte         `json:"signature"`
}

// GetFeeRateResponse is the json reponse for GetFeeRate
type GetFeeRateResponse struct {
	Result map[common.Hash]*FeeRateInfo `json:"result"`
}

// FeeRateInfo struct of fee-rate-info
type FeeRateInfo struct {
	ChannelID     common.Hash    `json:"channel_id"`
	PeerAddress   common.Address `json:"peer_address"`
	FeeRate       string         `json:"fee_rate"`
	EffectiveTime int64          `json:"effective_time"`
}

// SetFeeRate save request data of set_fee_rate
func SetFeeRate(req *http.Request, feeRateDB *storage.Database, peerAddress string) util.JSONResponse {

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

	_, err = strconv.ParseFloat(r.FeeRate, 32)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: util.InvalidArgumentValue("fee_rate must be a floating number(character string) from 0 to 1."),
		}
	}

	util.GetLogger(req.Context()).WithField("set_fee_rate", r.Signature).Info("Processing set_fee_rate request")

	var channelid = r.ChannelID
	var feerate = r.FeeRate
	err = feeRateDB.SaveRateFeeStorage(nil, channelid.String(), common.HexToAddress(peerAddress).String(), feerate)
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
func GetFeeRate(req *http.Request, feeRateDB *storage.Database, peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPost {
		var r GetFeeRateRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}

		err := verifySinatureGetFeeRate(r, common.HexToAddress(peerAddress))
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusBadRequest,
				JSON: err.Error(),
			}
		}

		feerate, effitime, err := feeRateDB.GetLastestRateFeeStorage(nil, r.ChannelID.String(), r.ObtainObj.String())
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusNotFound,
				JSON: util.NotFound("any fee-rate data found"),
			}
		}
		reslut0 := &FeeRateInfo{
			ChannelID:     r.ChannelID,
			PeerAddress:   common.HexToAddress(peerAddress),
			FeeRate:       feerate,
			EffectiveTime: effitime,
		}
		resultMap := make(map[common.Hash]*FeeRateInfo)
		resultMap[r.ChannelID] = reslut0
		reslut := &GetFeeRateResponse{
			Result: resultMap,
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
