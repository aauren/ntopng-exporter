package ntopng

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func getHttpResponseBody(client *http.Client, req *http.Request) (*[]byte, int, error) {
	var body []byte
	resp, err := client.Do(req) //nolint:gosec // URL is constructed from trusted application configuration, not user input
	if err != nil {
		return &body, 0, err
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
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

func (c *Controller) ResolveIfID(inputIfID int) (string, error) {
	for ifName, ifID := range c.ifList {
		if ifID == inputIfID {
			return ifName, nil
		}
	}
	return "", fmt.Errorf("could not find an interface name for ifid: %d", inputIfID)
}
