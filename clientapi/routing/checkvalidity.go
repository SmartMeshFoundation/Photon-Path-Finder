package routing

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"encoding/binary"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"fmt"
)

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