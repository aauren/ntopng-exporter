package ntopng

import (
	"bytes"
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"io/ioutil"
	"net/http"
)

const (
	luaRestV1Get = "/lua/rest/v1/get"
	hostCustomFields = `ip,bytes.sent,bytes.rcvd,active_flows.as_client,active_flows.as_server,dns,num_alerts,mac,
total_flows.as_client,total_flows.as_server,vlan,total_alerts,name`
	hostCustomPath = "/host/custom_data.lua"
)

type controller struct {
	config config.Config
}

func CreateController(config config.Config) controller {
	var controller controller
	controller.config = config
	return controller
}

func (c *controller) ScrapeHostEndpoint(interfaceId int) error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV1Get, hostCustomPath)
	payload := []byte(fmt.Sprintf(`{"ifid": %d, "field_alias": "%s"}`, interfaceId, hostCustomFields))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	c.setCommonOptions(req, true)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("response Status: %s\n", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("response Body: %s", string(body))
	return nil
}

func (c *controller) setCommonOptions(req *http.Request, isJsonRequest bool) {
	if isJsonRequest {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.config.Ntopng.AuthMethod == "cookie" {
		req.Header.Add("Cookie",
			fmt.Sprintf("user=%s; password=%s",
				c.config.Ntopng.User, c.config.Ntopng.Password))
	} else if c.config.Ntopng.AuthMethod == "basic" {
		req.SetBasicAuth(c.config.Ntopng.User, c.config.Ntopng.Password)
	}
}
