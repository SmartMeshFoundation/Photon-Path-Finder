package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"
)

func TestUpdateBalanceProof(t *testing.T) {
	bp := &BalanceProof{
		Nonce:             0,
		ChannelID:         common.StringToHash("0x4caea64ca26bce329e98faf70633581738dfc94cd0e1f5e1c4ad2bbb6386b63d"),
		TransferredAmount: big.NewInt(5),
		LocksRoot:         common.StringToHash("0x0585b5896ce265dc5221c6df8458b8c667686d3214e08b85c07e5ae194d26f5c"),
		AdditionalHash:    common.StringToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Signature:         common.StringToHash("0x0585b5896ce265dc5221c6df8458b8c667686d3214e08b85c07e5ae194d26f5c").Bytes(),
	}

	bpr := &balanceProofRequest{
		BalanceSignature: common.StringToHash("0x0000000000000000000000000000000000000000000000000000000000000000").Bytes(),
		BalanceProof:     *bp,
		LocksAmount:      big.NewInt(1),
	}

	httpurl := "http://localhost:9001/pathfinder/0xc67f23CE04ca5E8DD9f2E1B5eD4FaD877f79267A/balance"
	var reqBody interface{}
	reqBody = bpr
	var req *http.Request
	var err error
	httpclient := http.DefaultClient
	if reqBody != nil {
		var jsonStr []byte
		jsonStr, err = json.Marshal(reqBody)
		if err != nil {
			t.Errorf("Marshal json error: %s", err)
		}
		req, err = http.NewRequest("PUT", httpurl, bytes.NewBuffer(jsonStr))
		if err != nil {
			t.Errorf("error: %s", err)
		}
	} else {
		req, err = http.NewRequest("PUT", httpurl, nil)
		if err != nil {
			t.Errorf("error: %s", err)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	res, err := httpclient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		t.Errorf("error: %s", err)
	}
	_, err = ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		fmt.Println(fmt.Sprintf("UpdateBalanceProof error: %s", err))
	}
	fmt.Println("UpdateBalanceProof ok")

}
