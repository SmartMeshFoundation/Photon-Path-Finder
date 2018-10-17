package routing

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"encoding/binary"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"fmt"
)

// verifySinature verify balance proof sinature and caller's sinature
// 1\verify bob's balance proof sinature
// 2\verify alice(caller)'s infomation's sinature
// 3\Balance_Proof_Hash	(nonce,transferred_amount,channel_id,locksroot,additional_hash)
// 4\Message_Hash		(nonce,transferred_amount,channel_id,locksroot,additional_hash,locks_amount)
func verifySinature(bpr balanceProofRequest ,peerAddress common.Address) (partner common.Address,err error) {
	tmpBuf := new(bytes.Buffer)
	err = binary.Write(tmpBuf, binary.BigEndian, bpr.BalanceProof.Nonce)             //nonce
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferredAmount)) //transferred_amount
	_, err = tmpBuf.Write(bpr.BalanceProof.ChannelID[:])                             //channel_id
	_, err = tmpBuf.Write(bpr.BalanceProof.LocksRoot[:])                             //locksroot
	_, err = tmpBuf.Write(bpr.BalanceProof.AdditionalHash[:])                        //additional_hash

	balanceProofHash := utils.Sha3(tmpBuf.Bytes())
	balanceProofSignature := bpr.BalanceProof.Signature
	balanceProofSigner, err := utils.Ecrecover(balanceProofHash, balanceProofSignature)
	if balanceProofSigner == peerAddress {
		err = fmt.Errorf("Illegal balance proof signature")
		return balanceProofSigner, err
	}

	//The data size in communication may be very large.
	/*for _, lock := range bpr.Locks {
		err = binary.Write(tmpBuf, binary.BigEndian, lock.LockedAmount)
		err = binary.Write(tmpBuf, binary.BigEndian, lock.Expriation)
		_, err = tmpBuf.Write(lock.SecretHash[:])

		locksAmount.Add(locksAmount, lock.LockedAmount)
	}*/
	_, err = tmpBuf.Write(utils.BigIntTo32Bytes(bpr.LocksAmount)) //locks_amount

	messageHash := utils.Sha3(tmpBuf.Bytes())
	messageSignature := bpr.BalanceHash
	signer, err := utils.Ecrecover(messageHash, messageSignature[:])
	if signer != peerAddress {
		err = fmt.Errorf("Illegal signature of balance message")
		return balanceProofSigner, err
	}
	return balanceProofSigner, nil
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
	feeRateSigner, err := utils.Ecrecover(feeRateHash, feeRateHashSignature)
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
	pathSigner,err:=utils.Ecrecover(pathHash,pathSignature)
	if pathSigner!=peerAddress{
		err = fmt.Errorf("Invalid signature")
		return err
	}
	return nil
}