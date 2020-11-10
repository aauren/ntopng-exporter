package ntopng

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getHttpResponseBody(req *http.Request) (*[]byte, int, error) {
	var body []byte
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &body, 0, err
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	return &body, resp.StatusCode, nil
}

func getRawJsonFromNtopResponse(body *[]byte) (json.RawMessage, error) {
	var ntopResponse ntopResponse
	err := json.Unmarshal(*body, &ntopResponse)
	if err != nil {
		return nil, err
	}

	if ntopResponse.RcStr != "OK" {
		return nil, fmt.Errorf("interface response from ntopng was not successful. Response code: '%s'",
			ntopResponse.RcStr)
	}

	return ntopResponse.Rsp, nil
}
