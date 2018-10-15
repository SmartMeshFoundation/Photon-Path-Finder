package routing

import (
	"net/http"
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
)

//balanceProof is the json request for BalanceProof
type BalanceProof struct {
	Nonce             int64       `json:"nonce"`
	ChannelID         common.Hash `json:"channel_id"`
	TransferredAmount *big.Int    `json:"transferred_amount"`
	LocksRoot         common.Hash `json:"locksroot"`
	AdditionalHash    common.Hash `json:"additional_hash"`
	Signature         []byte      `json:"signature"`
}

//lock is the json request for BalanceProof
type lock struct {
	LockedAmount int         `json:"locked_amount"`
	Expriation   int         `json:"expiration"`
	SecretHash   common.Hash `json:"secret_hash"`
}

//balanceProofRequest is the json request for BalanceProof
type balanceProofRequest struct {
	BalanceHash  common.Hash  `json:"balance_hash"`
	BalanceProof BalanceProof `json:"balance_proof"`
	Locks        []lock       `json:"locks"`
}

// Balance handle the request with balance proof,implements GET and POST /balance
func UpdateBalanceProof(req *http.Request,cfg config.PathFinder,balanceDB *storage.Database, peerAddress string) util.JSONResponse {
	if req.Method == http.MethodPut {
		var r balanceProofRequest
		resErr := util.UnmarshalJSONRequest(req, &r)
		if resErr != nil {
			return *resErr
		}

		if resErr=r.Validate(peerAddress);resErr!=nil{
			return *resErr
		}

		util.GetLogger(req.Context()).WithField("balance_proof", r.BalanceHash).Info("Processing balance_proof request")


		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: util.InvalidArgumentValue("argument was incorrect"),
		}
		//write db
		/*return util.JSONResponse{
			Code: http.StatusOK,
			JSON: "ok",
		}*/
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: util.NotFound("Bad method"),
	}
}

func (r balanceProofRequest) Validate(peerAddressStr string) *util.JSONResponse {
	//验证request内容//判断signer地址//判断balance_hash//判断balance_proof中的signature
	//peerAddress := common.HexToAddress(peerAddressStr)
	//if peerAddress == utils.EmptyAddress {
	if common.IsHexAddress(peerAddressStr){
		return &util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: util.BadJSON("peerAddress must be provided"),
		}
	}
	//判断balance_hash

	//判断balance_proof中的signature
	return nil
}


