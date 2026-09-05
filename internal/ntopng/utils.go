package ntopng

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const maxLoggedResponse = 512

func getHttpResponseBody(client *http.Client, req *http.Request) (*[]byte, int, error) {
	var body []byte
	resp, err := client.Do(req)
	if err != nil {
		return &body, 0, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return &body, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}
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

// resolveIfNameOrID falls back to the numeric id so that an interface we can't name still gives us a
// usable metric label rather than an empty one
func (c *Controller) resolveIfNameOrID(inputIfID int) string {
	if ifName, err := c.ResolveIfID(inputIfID); err == nil {
		return ifName
	}
	return strconv.Itoa(inputIfID)
}

// responseSuffix appends ntopng's response to an error for debugging, truncated because a host list
// response can run to megabytes and we don't want that landing in the log on every failed scrape.
func responseSuffix(body *[]byte) string {
	if body == nil || len(*body) == 0 {
		return ""
	}
	if len(*body) > maxLoggedResponse {
		return fmt.Sprintf(", response: '%s' (truncated)", (*body)[:maxLoggedResponse])
	}
	return fmt.Sprintf(", response: '%s'", *body)
}
