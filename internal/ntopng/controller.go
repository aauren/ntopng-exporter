package ntopng

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"net/http"
	"strconv"
)

const (
	luaRestV1Get     = "/lua/rest/v1/get"
	hostCustomFields = `ip,bytes.sent,bytes.rcvd,active_flows.as_client,active_flows.as_server,dns,num_alerts,mac,total_flows.as_client,total_flows.as_server,vlan,total_alerts,name,ifid`
	hostCustomPath      = "/host/custom_data.lua"
	interfaceCustomPath = "/ntopng/interfaces.lua"
)

type Controller struct {
	config   config.Config
	ifList   map[string]int
	HostList map[string]ntopHost
}

func CreateController(config config.Config) Controller {
	var controller Controller
	controller.config = config
	return controller
}

func (c *Controller) CacheInterfaceIds() error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV1Get, interfaceCustomPath)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setCommonOptions(req, false)

	body, status, err := getHttpResponseBody(req)
	if status != http.StatusOK {
		return fmt.Errorf("request to interface endpoint was not successful. Status: '%d', Response: '%v'",
			status, *body)
	}

	rawInterfaces, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		return err
	}
	var ifList []ntopInterface
	err = json.Unmarshal(rawInterfaces, &ifList)
	if err != nil {
		return fmt.Errorf("was not able to parse interface list from ntopng: %v", err)
	}
	if len(ifList) < 1 {
		return fmt.Errorf("ntopng returned 0 interfaces: %v", *body)
	}
	c.ifList = make(map[string]int, len(ifList))
	for _, myIf := range ifList {
		c.ifList[myIf.IfName] = myIf.IfID
	}

	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		if _, ok := c.ifList[configuredIf]; !ok {
			return fmt.Errorf("could not find '%s' interface in list returned by ntopng: %v",
				configuredIf, c.ifList)
		}
	}
	return nil
}

func (c *Controller) ScrapeHostEndpointForAllInterfaces() error {
	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		if err := c.scrapeHostEndpoint(c.ifList[configuredIf]); err != nil {
			return fmt.Errorf("failed to scrape interface '%s' with error: %v", configuredIf, err)
		}
	}
	return nil
}

func (c *Controller) scrapeHostEndpoint(interfaceId int) error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV1Get, hostCustomPath)
	payload := []byte(fmt.Sprintf(`{"ifid": %d, "field_alias": "%s"}`, interfaceId, hostCustomFields))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	c.setCommonOptions(req, true)

	body, status, err := getHttpResponseBody(req)
	if status != http.StatusOK {
		return fmt.Errorf("request to host endpoint was not successful. Status: '%d', Response: '%v'",
			status, *body)
	}

	rawHosts, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		return err
	}
	var hostList []ntopHost
	err = json.Unmarshal(rawHosts, &hostList)
	if len(hostList) < 1 {
		return fmt.Errorf("ntopng returned 0 hosts: %v", *body)
	}
	if c.HostList == nil {
		c.HostList = make(map[string]ntopHost)
	}
	for _, myHost := range hostList {
		// If we already have this host in our cache and it has a different ifid than we are currently processing, don't
		// overwrite it, and print a warning.
		if err = c.checkForDuplicateInterfaces(&myHost); err != nil {
			fmt.Println(err)
			continue
		}
		if myHost.IfName, err = c.ResolveIfID(myHost.IfID); err != nil {
			fmt.Printf("Could not resolve interface: %d, this should not happen", myHost.IfID)
			myHost.IfName = strconv.Itoa(myHost.IfID)
		}
		c.HostList[myHost.IP] = myHost
	}
	return err
}

func (c *Controller) setCommonOptions(req *http.Request, isJsonRequest bool) {
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
