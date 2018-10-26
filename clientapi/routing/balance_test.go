package routing

import (
	"testing"
)

func TestUpdateBalanceProof(t *testing.T) {
	/*httpurl := "http://localhost:9001/pathfinder/0xc67f23CE04ca5E8DD9f2E1B5eD4FaD877f79267A/balance"
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
	fmt.Println("UpdateBalanceProof ok")*/

}
