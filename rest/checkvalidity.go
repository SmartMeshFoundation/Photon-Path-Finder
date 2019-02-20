package rest

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	smparams "github.com/SmartMeshFoundation/Photon/params"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
)

// verifyBalanceProofSignature verify balance proof sinature and caller's sinature
// 1\verify bob's balance proof sinature
// 2\verify alice(caller)'s infomation's sinature
// 3\Balance_Proof_Hash	(nonce,transfer_amount,locksroot,channel_id,open_block_number,additional_hash)
// 4\Message_Hash		(balance_proof,lock_amount)
func verifyBalanceProofSignature(bpr *balanceProofRequest, participant common.Address) (partner common.Address, err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.Nonce)           //nonce
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferAmount))  //transfer_amount
	_, err = tmpBuf.Write(bpr.BalanceProof.LocksRoot[:])                           //locksroot
	_, err = tmpBuf.Write(bpr.BalanceProof.ChannelID[:])                           //channel_id
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.OpenBlockNumber) //open_block_number
	_, err = tmpBuf.Write(bpr.BalanceProof.AdditionalHash[:])                      //additional_hash
	_, err = tmpBuf.Write(bpr.BalanceProof.Signature)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.LockedAmount)) //locks_amount
	_, err = tmpBuf.Write(bpr.ProofSigner[:])
	messageHash := utils.Sha3(tmpBuf.Bytes())
	messageSignature := bpr.BalanceSignature
	signer, err := utils.Ecrecover(messageHash, messageSignature)
	if signer != participant {
		err = fmt.Errorf("illegal signature of balance message, for participant")
		return
	}
	//ignore empty balance proof
	if bpr.BalanceProof.Nonce == 0 {
		partner = bpr.ProofSigner
		return
	}
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
	partner, err = utils.Ecrecover(balanceProofHash, balanceProofSignature)
	if err != nil {
		err = fmt.Errorf("illegal balance proof signature %s", err)
		return
	}
	return
}

// verifySinatureSetFeeRate verify Fee_rate sinature
func verifySinatureSetFeeRate(sfr *SetFeeRateRequest, peerAddress common.Address) (err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, sfr.FeePercent)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(sfr.FeeConstant))

	feeRateHash := utils.Sha3(tmpBuf.Bytes())
	feeRateHashSignature := sfr.Signature
	feeRateSigner, err := utils.Ecrecover(feeRateHash, feeRateHashSignature)
	if feeRateSigner != peerAddress {
		err = fmt.Errorf("invalid signature of set fee_rate")
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
	pathSignature := pr.Signature
	pathSigner, err := utils.Ecrecover(pathHash, pathSignature)
	if pathSigner != peerAddress {
		err = fmt.Errorf("Invalid signature")
		return err
	}
	return nil
}

// SignDataForBalanceProof0 signature data,just for test
func SignDataForBalanceProof0(peerKey *ecdsa.PrivateKey, bp *model.BalanceProof) (err error) {
	bpBuf := new(bytes.Buffer)
	_, err = bpBuf.Write(smparams.ContractSignaturePrefix)
	_, err = bpBuf.Write([]byte(smparams.ContractBalanceProofMessageLength))
	_, err = bpBuf.Write(utils.BigIntTo32Bytes(bp.TransferAmount))
	_, err = bpBuf.Write(bp.LocksRoot[:])
	err = binary.Write(bpBuf, binary.BigEndian, bp.Nonce)
	_, err = bpBuf.Write(bp.AdditionalHash[:])
	_, err = bpBuf.Write(bp.ChannelID[:])
	err = binary.Write(bpBuf, binary.BigEndian, bp.OpenBlockNumber)
	_, err = bpBuf.Write(utils.BigIntTo32Bytes(params.ChainID)) //smparams.ChainID
	bp.Signature, err = utils.SignData(peerKey, bpBuf.Bytes())

	return
}

// SignDataForBalanceProofMessage0 signature data(balance proof),just for test
func SignDataForBalanceProofMessage0(peerKey *ecdsa.PrivateKey, r *balanceProofRequest) (err error) {

	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, r.BalanceProof.Nonce)
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.BalanceProof.TransferAmount))
	_, err = tmpBuf.Write(r.BalanceProof.LocksRoot[:])
	_, err = tmpBuf.Write(r.BalanceProof.ChannelID[:])
	err = binary.Write(tmpBuf, binary.BigEndian, r.BalanceProof.OpenBlockNumber)
	_, err = tmpBuf.Write(r.BalanceProof.AdditionalHash[:])
	_, err = tmpBuf.Write(r.BalanceProof.Signature[:])
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(r.LockedAmount))
	r.BalanceSignature, err = utils.SignData(peerKey, tmpBuf.Bytes())
	return
}
