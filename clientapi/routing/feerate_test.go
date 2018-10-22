package routing

import (
	"testing"
	"net/http"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bytes"
)

func TestSetFeeRate(t *testing.T) {
	channelid := "0x4caea64ca26bce329e98faf70633581738dfc94cd0e1f5e1c4ad2bbb6386b63d"
	feerate0 := "0.001"
	feerate1 := "~!@#$%^&*()_+`1234567890-=abc"
	signature0 := "v2mCuIM6rpLZ1IRlxiiYL79JWOABKjctrSTUiekVAe82XN/GKj6nxxVQSwIdoygUms2dlWnKJQO63RAuEAuXuRs="
	signature1 := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	peeraddress := "0xc67f23ce04ca5e8dd9f2e1b5ed4fad877f79267a"
	httpUrl := "http://127..1:9001/pathfinder/" + peeraddress + "/set_fee_rate"
	/*var req *http.Request
	var feeRateDB *storage.Database*/

	request0 := &SetFeeRateRequest{
		ChannelID: common.HexToHash(channelid),
		FeeRate:   feerate0,
		Signature: []byte(signature0),
	}
	request1 := &SetFeeRateRequest{
		ChannelID: common.HexToHash(channelid),
		FeeRate:   feerate1,
		Signature: []byte(signature1),
	}
	var response0 *util.JSONResponse
	var response1 *util.JSONResponse
	_, err0 := MakeRequest("PUT", httpUrl, request0, response0)
	_, err1 := MakeRequest("PUT", httpUrl, request1, response1)
	fmt.Println(err0,response0)
	fmt.Println(err1,response1)
}

func MakeRequest(method string, httpURL string, reqBody interface{}, resBody interface{}) ([]byte, error) {
	client := &http.Client{}
	var req *http.Request
	var err error
	if reqBody != nil {
		var jsonStr []byte
		jsonStr, err = json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, httpURL, bytes.NewBuffer(jsonStr))
	} else {
		req, err = http.NewRequest(method, httpURL, nil)
	}
	//log.Trace(fmt.Sprintf("matrix url:%s,req:%s", httpURL, reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 != 2 {
		var wrap error
		var respErr RespError
		err = json.Unmarshal(contents, &respErr)
		if err != nil {
			return nil, err
		}
		if respErr.ErrCode != "" {
			wrap = respErr
		}
		msg := "Failed to " + method + " JSON to " + req.URL.Path
		if wrap == nil {
			msg = msg + ": " + string(contents)
		}
		return contents, HTTPError{
			Code:         res.StatusCode,
			Message:      msg,
			WrappedError: wrap,
		}
	}
	if err != nil {
		return nil, err
	}
	if resBody != nil {
		if err = json.Unmarshal(contents, &resBody); err != nil {
			return nil, err
		}
	}
	return contents, nil
}

type RespError struct {
	ErrCode string `json:"errcode"`
	Err     string `json:"error"`
}

type HTTPError struct {
	WrappedError error
	Message      string
	Code         int
}

func (e HTTPError) Error() string {
	var wrappedErrMsg string
	if e.WrappedError != nil {
		wrappedErrMsg = e.WrappedError.Error()
	}
	return fmt.Sprintf("msg=%s code=%d wrapped=%s", e.Message, e.Code, wrappedErrMsg)
}

func (e RespError) Error() string {
	return e.ErrCode + ": " + e.Err
}