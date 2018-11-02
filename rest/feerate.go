package rest

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/SmartMeshFoundation/Photon/log"
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
		}
		policy = model.FeePolicyCombined
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
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = tn.UpdateChannelFeeRate(channel, peerAddress, fee)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}

// getChannelFeeRate reponse fee_rate data
func getChannelRate(w rest.ResponseWriter, r *rest.Request) {

	peerAddress := common.HexToAddress(r.PathParam("peer"))
	channelID := common.HexToHash(r.PathParam("channel"))
	fee, err := model.GetChannelFeeRate(channelID, peerAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
	return
}

func setTokenRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))
	token := common.HexToAddress(r.PathParam("token"))
	var req SetFeeRateRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = model.UpdateAccountTokenFee(peerAddress, token, fee)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}
func getTokenRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))
	token := common.HexToAddress(r.PathParam("token"))
	fee, err := model.GetAccountTokenFee(peerAddress, token)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
	return
}
func setAccountRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))

	var req SetFeeRateRequest
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//validate json-input
	err = verifySinatureSetFeeRate(&req, peerAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	policy, err := verifyAndGetFeePolicy(&req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fee := &model.Fee{
		FeePolicy:   policy,
		FeePercent:  req.FeePercent,
		FeeConstant: req.FeeConstant,
	}
	err = model.UpdateAccountDefaultFeePolicy(peerAddress, fee)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
}
func getAccountRate(w rest.ResponseWriter, r *rest.Request) {
	peerAddress := common.HexToAddress(r.PathParam("peer"))

	fee := model.GetAccountFeePolicy(peerAddress)

	err := w.WriteJson(fee)
	if err != nil {
		log.Error(fmt.Sprintf("write json err %s", err))
	}
	return
}
