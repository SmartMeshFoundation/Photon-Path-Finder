package routing

import (
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"encoding/binary"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"fmt"
	"math/big"
)

func verifySinature(bpr balanceProofRequest ,peerAddress common.Address) (partner common.Address,locksAmount *big.Int,err error){
	tmpBuf:=new(bytes.Buffer)
	err=binary.Write(tmpBuf,binary.BigEndian,bpr.BalanceProof.Nonce)
	_,err=tmpBuf.Write(utils.BigIntTo32Bytes(bpr.BalanceProof.TransferredAmount))
	_, err = tmpBuf.Write(bpr.BalanceProof.LocksRoot[:])
	_, err = tmpBuf.Write(bpr.BalanceProof.AdditionalHash[:])
	balanceProofHash:=utils.Sha3(tmpBuf.Bytes())
	balanceProofSignature:=bpr.BalanceProof.Signature
	balanceProofSigner,err:=utils.Ecrecover(balanceProofHash,balanceProofSignature)
	if balanceProofSigner==peerAddress{
		err=fmt.Errorf("Illegal balance proof signature")
		return balanceProofSigner,big.NewInt(0),err
	}

	for _,lock:=range bpr.Locks{
		err=binary.Write(tmpBuf,binary.BigEndian,lock.LockedAmount)
		err=binary.Write(tmpBuf,binary.BigEndian,lock.Expriation)
		_, err = tmpBuf.Write(lock.SecretHash[:])

		locksAmount.Add(locksAmount,lock.LockedAmount)
	}
	messageHash:=utils.Sha3(tmpBuf.Bytes())
	messageSignature:=bpr.BalanceHash
	signer,err:=utils.Ecrecover(messageHash,messageSignature[:])
	if signer!=peerAddress{
		err=fmt.Errorf("Illegal signature of balance message")
		return balanceProofSigner,big.NewInt(0),err
	}
	return balanceProofSigner,locksAmount,nil
}