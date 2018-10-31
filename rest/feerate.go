package rest

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/ethereum/go-ethereum/common"
)

// SetFeeRateRequest is the json request for setChannelRate
type SetFeeRateRequest struct {
	FeeConstant *big.Int `json:"fee_constant"`
	FeePercent  int64    `json:"fee_percent"`
	Signature   []byte   `json:"signature"`
}

func verifyAndGetFeePolicy(req *SetFeeRateRequest) (policy int, err error) {
	policy = model.FeePolicyConstant
	if req.FeePercent > 0 {
		policy = model.FeePolicyPercent
		if req.FeeConstant != nil && req.FeeConstant.Cmp(utils.BigInt0) > 0 {
			policy = model.FeePolicyCombined
		}
	} else {
		policy = model.FeePolicyConstant
		if req.FeeConstant == nil || req.FeeConstant.Cmp(utils.BigInt0) < 0 {
			err = fmt.Errorf("fee arg err constant=%s,percent=%d", req.FeeConstant, req.FeePercent)
			return
		} else {
			policy = model.FeePolicyCombined
		}
	}
	return
}

// setChannelRate save request data of set_fee_rate
func setChannelRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))
	channel := common.HexToHash(r.PathParam("channel"))
	var req SetFeeRateRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = tn.UpdateChannelFeeRate(channel, peerAddress, fee)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
}

// getChannelFeeRate reponse fee_rate data
func getChannelRate(w rest.ResponseWriter, r *rest.Request) {

	peerAddress := common.HexToAddress(r.PathParam("peer"))
	channelID := common.HexToHash(r.PathParam("channel"))
	fee, err := model.GetChannelFeeRate(channelID, peerAddress)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}

	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
	return
}

func setTokenRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))
	token := common.HexToAddress(r.PathParam("token"))
	var req SetFeeRateRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = model.UpdateAccountTokenFee(peerAddress, token, fee)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
}
func getTokenRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))
	token := common.HexToAddress(r.PathParam("token"))
	fee, err := model.GetAccountTokenFee(peerAddress, token)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}

	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
	return
}
func setAccountRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))

	var req SetFeeRateRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = model.UpdateAccountDefaultFeePolicy(peerAddress, fee)
	if err != nil {
		w.WriteJson(&response{
			Code: http.StatusBadRequest,
			JSON: err.Error(),
		})
		return
	}
	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
}
func getAccountRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))

	fee := model.GetAccountFeePolicy(peerAddress)

	w.WriteJson(&response{
		Code: http.StatusOK,
		JSON: fee,
	})
	return
}
