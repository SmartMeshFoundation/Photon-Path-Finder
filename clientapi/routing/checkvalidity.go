package routing

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/params"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden/accounts"
	smparams "github.com/SmartMeshFoundation/SmartRaiden/params"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// verifySinature verify balance proof sinature and caller's sinature
// 1\verify bob's balance proof sinature
// 2\verify alice(caller)'s infomation's sinature
// 3\Balance_Proof_Hash	(nonce,transfer_amount,locksroot,channel_id,open_block_number,additional_hash)
// 4\Message_Hash		(balance_proof,lock_amount)
func verifySinature(bpr *balanceProofRequest, peerAddress common.Address, partner common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.Nonce)           //nonce
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferAmount))  //transfer_amount
	_, err = tmpBuf.Write(bpr.BalanceProof.LocksRoot[:])                           //locksroot
	_, err = tmpBuf.Write(bpr.BalanceProof.ChannelID[:])                           //channel_id
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.OpenBlockNumber) //open_block_number
	_, err = tmpBuf.Write(bpr.BalanceProof.AdditionalHash[:])                      //additional_hash

	//检查是谁的balance proof
	bpBuf := new(bytes.Buffer)
	_, err = bpBuf.Write(smparams.ContractSignaturePrefix)
	_, err = bpBuf.Write([]byte(smparams.ContractBalanceProofMessageLength))
	_, err = bpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferAmount))
	_, err = bpBuf.Write(bpr.BalanceProof.LocksRoot[:])
	err = binary.Write(bpBuf, binary.BigEndian, bpr.BalanceProof.Nonce)
	_, err = bpBuf.Write(bpr.BalanceProof.AdditionalHash[:])
	_, err = bpBuf.Write(bpr.BalanceProof.ChannelID[:])
	err = binary.Write(bpBuf, binary.BigEndian, bpr.BalanceProof.OpenBlockNumber)
	_, err = bpBuf.Write(utils.BigIntTo32Bytes(params.ChainID)) //smparams.ChainID
	balanceProofHash := utils.Sha3(bpBuf.Bytes())
	balanceProofSignature := bpr.BalanceProof.Signature
	balanceProofSigner, err := utils.Ecrecover(balanceProofHash, balanceProofSignature)
	if err != nil {
		err = fmt.Errorf("Illegal balance proof signature")
		return err
	}
	if balanceProofSigner != partner {
		err = fmt.Errorf("Illegal balance proof signature,must give partner's balance proof")
		return err
	}

	_, err = tmpBuf.Write(bpr.BalanceProof.Signature)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.LocksAmount)) //locks_amount
	messageHash := utils.Sha3(tmpBuf.Bytes())
	messageSignature := bpr.BalanceSignature
	signer, err := utils.Ecrecover(messageHash, messageSignature)
	if signer != peerAddress {
		err = fmt.Errorf("Illegal signature of balance message")
		return err
	}
	return nil
}

// verifySinatureSetFeeRate verify Fee_rate sinature
func verifySinatureSetFeeRate(sfr SetFeeRateRequest, peerAddress common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	_, err = tmpBuf.Write(sfr.ChannelID[:])    //channel_id
	_, err = tmpBuf.Write([]byte(sfr.FeeRate)) //fee_rate

	feeRateHash := utils.Sha3(tmpBuf.Bytes())
	feeRateHashSignature := sfr.Signature
	feeRateSigner, err := utils.Ecrecover(feeRateHash, feeRateHashSignature)
	if feeRateSigner != peerAddress {
		err = fmt.Errorf("Invalid signature of set fee_rate")
		return err
	}
	return nil
}

// verifySinatureGetFeeRate verify Fee_rate sinature
func verifySinatureGetFeeRate(sfr GetFeeRateRequest, peerAddress common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	_, err = tmpBuf.Write(sfr.ObtainObj[:]) //obtain_obj
	_, err = tmpBuf.Write(sfr.ChannelID[:]) //channel_id

	feeRateHash := utils.Sha3(tmpBuf.Bytes())
	feeRateHashSignature := sfr.Signature
	feeRateSigner, err := utils.Ecrecover(feeRateHash, feeRateHashSignature)
	if feeRateSigner != peerAddress {
		err = fmt.Errorf("Invalid signature of get fee_rate")
		return err
	}
	return nil
}

//verifySinaturePaths signature=caller
func verifySinaturePaths(pr pathRequest, peerAddress common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	_, err = tmpBuf.Write(pr.PeerFrom[:])                       //peer_from
	_, err = tmpBuf.Write(pr.PeerTo[:])                         //peer_to
	_, err = tmpBuf.Write(pr.TokenAddress[:])                   //token_address
	err = binary.Write(tmpBuf, binary.BigEndian, pr.LimitPaths) //limit_paths
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(pr.SendAmount)) //send_amount
	_, err = tmpBuf.Write([]byte(pr.SortDemand))                //send_amount

	pathHash := utils.Sha3(tmpBuf.Bytes())
	pathSignature := pr.Sinature
	pathSigner, err := utils.Ecrecover(pathHash, pathSignature)
	if pathSigner != peerAddress {
		err = fmt.Errorf("Invalid signature")
		return err
	}
	return nil
}

// SignData signature data,just for test
func signDataForPath(req *http.Request, cfg config.PathFinder, peerAddress string) util.JSONResponse {
	if req.Method != http.MethodPost {
		return util.JSONResponse{
			Code: http.StatusMethodNotAllowed,
			JSON: util.NotFound("Bad method"),
		}
	}
	var r pathRequest
	resErr := util.UnmarshalJSONRequest(req, &r)
	if resErr != nil {
		return *resErr
	}
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	_, err := tmpBuf.Write(r.PeerFrom[:])
	_, err = tmpBuf.Write(r.PeerTo[:])
	_, err = tmpBuf.Write(r.TokenAddress[:])
	err = binary.Write(tmpBuf, binary.BigEndian, r.LimitPaths)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.SendAmount))
	_, err = tmpBuf.Write([]byte(r.SortDemand))
	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey, err := crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey, tmpBuf.Bytes())
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: signature,
	}
}

// SignData signature data(balance proof),just for test
func signDataForSetFee(req *http.Request, cfg config.PathFinder, peerAddress string) util.JSONResponse {
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
	_, err := strconv.ParseFloat(r.FeeRate, 32)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: util.InvalidArgumentValue("fee_rate must be a floating number(character string) from 0 to 1."),
		}
	}
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	_, err = tmpBuf.Write(r.ChannelID[:])
	_, err = tmpBuf.Write([]byte(r.FeeRate))

	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey, err := crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey, tmpBuf.Bytes())
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: signature,
	}
}

// signDataForGetFee signature data(balance proof),just for test
func signDataForGetFee(req *http.Request, cfg config.PathFinder, peerAddress string) util.JSONResponse {
	if req.Method != http.MethodPost {
		return util.JSONResponse{
			Code: http.StatusMethodNotAllowed,
			JSON: util.NotFound("Bad method"),
		}
	}
	var r GetFeeRateRequest
	resErr := util.UnmarshalJSONRequest(req, &r)
	if resErr != nil {
		return *resErr
	}
	var signature []byte
	tmpBuf := new(bytes.Buffer)
	_, err := tmpBuf.Write(r.ObtainObj[:])
	_, err = tmpBuf.Write(r.ChannelID[:])

	accmanager := accounts.NewAccountManager(cfg.KeystorePath)
	privkeybin, err := accmanager.GetPrivateKey(common.HexToAddress(peerAddress), "123")
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	privateKey, err := crypto.ToECDSA(privkeybin)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: err.Error(),
		}
	}
	signature, err = utils.SignData(privateKey, tmpBuf.Bytes())
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusExpectationFailed,
			JSON: util.NotFound("Sign data err"),
		}
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: signature,
	}
}

// SignDataForBalanceProof0 signature data,just for test
func SignDataForBalanceProof0(peerKey *ecdsa.PrivateKey, r *BalanceProof) (err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, r.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.TransferAmount))
	_, err = tmpBuf.Write(r.ChannelID[:])
	_, err = tmpBuf.Write(r.LocksRoot[:])
	_, err = tmpBuf.Write(r.AdditionalHash[:])
	r.Signature, err = utils.SignData(peerKey, tmpBuf.Bytes())

	return
}

// SignDataForBalanceProofMessage0 signature data(balance proof),just for test
func SignDataForBalanceProofMessage0(peerKey *ecdsa.PrivateKey, r *balanceProofRequest) (err error) {

	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, r.BalanceProof.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.BalanceProof.TransferAmount))
	_, err = tmpBuf.Write(r.BalanceProof.ChannelID[:])
	_, err = tmpBuf.Write(r.BalanceProof.LocksRoot[:])
	_, err = tmpBuf.Write(r.BalanceProof.AdditionalHash[:])
	_, err = tmpBuf.Write(r.BalanceProof.Signature[:])
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.LocksAmount))
	r.BalanceSignature, err = utils.SignData(peerKey, tmpBuf.Bytes())
	return
}
