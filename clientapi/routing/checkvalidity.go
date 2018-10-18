package routing

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"encoding/binary"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"fmt"
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"crypto/ecdsa"
)

// verifySinature verify balance proof sinature and caller's sinature
// 1\verify bob's balance proof sinature
// 2\verify alice(caller)'s infomation's sinature
// 3\Balance_Proof_Hash	(nonce,transferred_amount,channel_id,locksroot,additional_hash)
// 4\Message_Hash		(nonce,transferred_amount,channel_id,locksroot,additional_hash,locks_amount)
func verifySinature(bpr * balanceProofRequest ,peerAddress common.Address,partner common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.Nonce)             //nonce
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferredAmount)) //transferred_amount
	_, err = tmpBuf.Write(bpr.BalanceProof.ChannelID[:])                             //channel_id
	_, err = tmpBuf.Write(bpr.BalanceProof.LocksRoot[:])                             //locksroot
	_, err = tmpBuf.Write(bpr.BalanceProof.AdditionalHash[:])                        //additional_hash

	balanceProofHash := utils.Sha3(tmpBuf.Bytes())
	balanceProofSignature := bpr.BalanceProof.Signature
	balanceProofSigner, err := utils.Ecrecover(balanceProofHash, balanceProofSignature)
	if err!=nil{
		err = fmt.Errorf("Illegal balance proof signature")
		return  err
	}
	if balanceProofSigner != partner {
		err = fmt.Errorf("Illegal balance proof signature,must give partner's balance proof")
		return  err
	}

	//The data size in communication may be very large.
	/*for _, lock := range bpr.Locks {
		err = binary.Write(tmpBuf, binary.BigEndian, lock.LockedAmount)
		err = binary.Write(tmpBuf, binary.BigEndian, lock.Expriation)
		_, err = tmpBuf.Write(lock.SecretHash[:])

		locksAmount.Add(locksAmount, lock.LockedAmount)
	}*/
	_,err=tmpBuf.Write(bpr.BalanceProof.Signature)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.LocksAmount)) //locks_amount

	messageHash := utils.Sha3(tmpBuf.Bytes())
	messageSignature := bpr.BalanceSignature
	signer, err := utils.Ecrecover(messageHash, messageSignature)
	if signer != peerAddress {
		err = fmt.Errorf("Illegal signature of balance message")
		return  err
	}
	return  nil
}

// verifySinatureFeeRate verify Fee_rate sinature
// 1\verify alice(caller)'s sinature
// 2\Balance_Proof_Hash	(channel_id,fee_rate)
func verifySinatureFeeRate(sfr SetFeeRateRequest ,peerAddress common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	_, err = tmpBuf.Write(sfr.ChannelID[:])                   //channel_id
	err = binary.Write(tmpBuf, binary.BigEndian, sfr.FeeRate) //fee_rate

	feeRateHash := utils.Sha3(tmpBuf.Bytes())
	feeRateHashSignature := sfr.Signature
	feeRateSigner, err := utils.Ecrecover(feeRateHash, feeRateHashSignature.Bytes())
	if feeRateSigner != peerAddress {
		err = fmt.Errorf("Invalid signature")
		return err
	}
	return nil
}

//verifySinaturePaths signature=caller
func verifySinaturePaths(pr pathRequest,peerAddress common.Address) (err error)  {
	tmpBuf := new(bytes.Buffer)
	_,err=tmpBuf.Write(pr.PeerFrom[:])//peer_from
	_,err=tmpBuf.Write(pr.PeerTo[:])//peer_to
	err = binary.Write(tmpBuf, binary.BigEndian, pr.LimitPaths) //limit_paths
	err = binary.Write(tmpBuf, binary.BigEndian, pr.SendAmount) //send_amount
	err = binary.Write(tmpBuf, binary.BigEndian, pr.SortDemand) //send_amount

	pathHash:=utils.Sha3(tmpBuf.Bytes())
	pathSignature:=pr.Sinature
	pathSigner,err:=utils.Ecrecover(pathHash,pathSignature.Bytes())
	if pathSigner!=peerAddress{
		err = fmt.Errorf("Invalid signature")
		return err
	}
	return nil
}

// SignData signature data,just for test
func signDataForBalanceProof(req *http.Request,cfg config.PathFinder,peerAddress string)  util.JSONResponse {
	if req.Method != http.MethodPost {
		return util.JSONResponse{
			Code: http.StatusMethodNotAllowed,
			JSON: util.NotFound("Bad method"),
		}
	}
	var r BalanceProof
	resErr := util.UnmarshalJSONRequest(req, &r)
	if resErr != nil {
		return *resErr
	}
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	err := binary.Write(tmpBuf, binary.BigEndian, r.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.TransferredAmount))
	_, err = tmpBuf.Write(r.ChannelID[:])
	_, err = tmpBuf.Write(r.LocksRoot[:])
	_, err = tmpBuf.Write(r.AdditionalHash[:])
	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey,err:=crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey,tmpBuf.Bytes())
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: common.BytesToHash(signature).String(),
	}
}

// SignData signature data,just for test
func signDataForBalanceProofMessage(req *http.Request,cfg config.PathFinder,peerAddress string)  util.JSONResponse {
	if req.Method != http.MethodPost {
		return util.JSONResponse{
			Code: http.StatusMethodNotAllowed,
			JSON: util.NotFound("Bad method"),
		}
	}
	var r balanceProofRequest
	resErr := util.UnmarshalJSONRequest(req, &r)
	if resErr != nil {
		return *resErr
	}
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	err := binary.Write(tmpBuf, binary.BigEndian, r.BalanceProof.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.BalanceProof.TransferredAmount))
	_, err = tmpBuf.Write(r.BalanceProof.ChannelID[:])
	_, err = tmpBuf.Write(r.BalanceProof.LocksRoot[:])
	_, err = tmpBuf.Write(r.BalanceProof.AdditionalHash[:])
	_, err = tmpBuf.Write(r.BalanceProof.Signature[:])
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.LocksAmount))
	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey,err:=crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey,tmpBuf.Bytes())
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: common.BytesToHash(signature).String(),
	}
}

// SignData signature data(balance proof),just for test
func signDataForSetFee(req *http.Request,cfg config.PathFinder,peerAddress string)  util.JSONResponse {
	if req.Method != http.MethodPost {
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
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	_, err:= tmpBuf.Write(r.ChannelID[:])
	_, err = tmpBuf.Write(utils.StringToBytes(r.FeeRate))

	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey,err:=crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey,tmpBuf.Bytes())
	if err!=nil{
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: common.BytesToHash(signature).String(),
	}
}

// SignData signature data,just for test
func SignDataForBalanceProof0(peerKey *ecdsa.PrivateKey,r *BalanceProof)  (err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, r.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.TransferredAmount))
	_, err = tmpBuf.Write(r.ChannelID[:])
	_, err = tmpBuf.Write(r.LocksRoot[:])
	_, err = tmpBuf.Write(r.AdditionalHash[:])
	r.Signature, err = utils.SignData(peerKey,tmpBuf.Bytes())

	return
}

// SignData signature data(balance proof),just for test
func SignDataForBalanceProofMessage0(peerKey *ecdsa.PrivateKey,r *balanceProofRequest)  (err error) {

	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, r.BalanceProof.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.BalanceProof.TransferredAmount))
	_, err = tmpBuf.Write(r.BalanceProof.ChannelID[:])
	_, err = tmpBuf.Write(r.BalanceProof.LocksRoot[:])
	_, err = tmpBuf.Write(r.BalanceProof.AdditionalHash[:])
	_, err = tmpBuf.Write(r.BalanceProof.Signature[:])
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.LocksAmount))
	r.BalanceSignature, err = utils.SignData(peerKey, tmpBuf.Bytes())
	return
}