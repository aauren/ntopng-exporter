package ntopng

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aauren/ntopng-exporter/internal"
	"github.com/aauren/ntopng-exporter/internal/config"
)

const (
	luaRestV2Get     = "/lua/rest/v2/get"
	hostCustomFields = `ip,bytes.sent,bytes.rcvd,active_flows.as_client,active_flows.as_server,dns,` +
		`num_alerts,mac,total_flows.as_client,total_flows.as_server,vlan,total_alerts,name,ifid,` +
		`packets.rcvd,packets.sent`
	hostCustomPath    = "/host/custom_data.lua"
	interfaceListPath = "/ntopng/interfaces.lua"
	interfaceDataPath = "/interface/data.lua"
	l7DataPath        = "/interface/l7/data.lua"
)

type Controller struct {
	config         *config.Config
	ifList         map[string]int
	HostList       map[hostKey]ntopHost
	InterfaceList  map[string]ntopInterfaceFull
	L7ProtocolList map[string]ntopL7Protocol
	ListRWMutex    *sync.RWMutex
	stopChan       <-chan struct{}
}

func CreateController(config *config.Config, stopChan <-chan struct{}) Controller {
	var controller Controller
	controller.config = config
	controller.stopChan = stopChan
	controller.ListRWMutex = &sync.RWMutex{}
	return controller
}

func (c *Controller) RunController() {
	scrapeInterval, err := time.ParseDuration(c.config.Ntopng.ScrapeInterval)
	if err != nil {
		fmt.Printf("was not able to parse duration: %s - %v", c.config.Ntopng.ScrapeInterval, err)
		return
	}
	ticker := time.NewTicker(scrapeInterval)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("scrap interval hit: scraping from ntop\n")
			c.ScrapeAllConfiguredTargets()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Controller) ScrapeAllConfiguredTargets() {
	if internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.HostScrape) ||
		internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.AllScrape) {
		c.ScrapeHostEndpointForAllInterfaces()
	}
	if internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.InterfaceScrape) ||
		internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.AllScrape) {
		c.ScrapeInterfaceEndpointForAllInterfaces()
	}
	if internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.L7Protocols) ||
		internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.AllScrape) {
		c.ScrapeL7ProtocolEndpointForAllInterfaces()
	}
}

func (c *Controller) CacheInterfaceIds() error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV2Get, interfaceListPath)
	req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to get response from ntopng interface endpoint: %v", err)
	}
	c.setCommonOptions(req, false)

	body, status, _ := getHttpResponseBody(getHttpClient(c.config.Ntopng.AllowUnsafeTLS), req)
	if status != http.StatusOK {
		if body != nil {
			return fmt.Errorf("request to interface endpoint was not successful. Status: '%d', Response: '%v'",
				status, *body)
		} else {
			return fmt.Errorf("request to interface endpoint was not successful. Status: '%d'",
				status)
		}
	}

	rawInterfaces, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		fmt.Printf("Received the following HTTP body response when we were expecting JSON: \n%s\n", body)
		return fmt.Errorf("failed to parse JSON from HTTP body: %v", err)
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

func (c *Controller) ScrapeHostEndpointForAllInterfaces() {
	// tempNtopHosts is made here to minimize the amount of time we have to lock the list and also to make sure that we
	// don't keep a list of ever growing hosts in our map which could eventually overwhelm the system
	tempNtopHosts := make(map[hostKey]ntopHost)
	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		if err := c.scrapeHostEndpoint(c.ifList[configuredIf], tempNtopHosts); err != nil {
			fmt.Printf("failed to scrape interface '%s' with error: %v", configuredIf, err)
		}
	}
	c.ListRWMutex.Lock()
	defer c.ListRWMutex.Unlock()
	c.HostList = tempNtopHosts
}

func (c *Controller) scrapeHostEndpoint(interfaceId int, tempNtopHosts map[hostKey]ntopHost) error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV2Get, hostCustomPath)
	payload := []byte(fmt.Sprintf(`{"ifid": %d, "field_alias": "%s"}`, interfaceId, hostCustomFields))
	req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	c.setCommonOptions(req, true)

	body, status, _ := getHttpResponseBody(getHttpClient(c.config.Ntopng.AllowUnsafeTLS), req)
	if status != http.StatusOK {
		if body != nil {
			return fmt.Errorf("request to host endpoint was not successful. Status: '%d', Response: '%v'",
				status, *body)
		} else {
			return fmt.Errorf("request to host endpoint was not successful. Status: '%d'",
				status)
		}
	}

	rawHosts, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		return err
	}
	var hostList []ntopHost
	_ = json.Unmarshal(rawHosts, &hostList)
	if len(hostList) < 1 {
		return fmt.Errorf("ntopng returned 0 hosts: %v", *body)
	}
	var parsedSubnets []*net.IPNet
	if len(c.config.Metric.LocalSubnetsOnly) > 0 {
		for _, subnet := range c.config.Metric.LocalSubnetsOnly {
			_, parsedSubnet, _ := net.ParseCIDR(subnet)
			parsedSubnets = append(parsedSubnets, parsedSubnet)
		}
	}
	for _, myHost := range hostList {
		if len(parsedSubnets) > 0 {
			validIP := false
			parsedIP := net.ParseIP(myHost.IP)
			for _, parsedSubnet := range parsedSubnets {
				if parsedSubnet.Contains(parsedIP) {
					validIP = true
					break
				}
			}
			if !validIP {
				continue
			}
		}
		if myHost.IfName, err = c.ResolveIfID(myHost.IfID); err != nil {
			fmt.Printf("Could not resolve interface: %d, this should not happen", myHost.IfID)
			myHost.IfName = strconv.Itoa(myHost.IfID)
		}
		tempNtopHosts[hostKey{IP: myHost.IP, IfID: myHost.IfID, VLAN: myHost.VLAN}] = myHost
	}
	return nil
}

func (c *Controller) ScrapeInterfaceEndpointForAllInterfaces() {
	// tempNtopInterfaces is made here to minimize the amount of time we have to lock the list and also to make sure that we
	// don't keep a list of ever growing hosts in our map which could eventually overwhelm the system
	tempNtopInterfaces := make(map[string]ntopInterfaceFull)
	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		if err := c.scrapeInterfaceEndpoint(c.ifList[configuredIf], tempNtopInterfaces); err != nil {
			fmt.Printf("failed to scrape interface '%s' with error: %v", configuredIf, err)
		}
	}
	c.ListRWMutex.Lock()
	defer c.ListRWMutex.Unlock()
	c.InterfaceList = tempNtopInterfaces
}

func (c *Controller) scrapeInterfaceEndpoint(interfaceId int, tempInterfaces map[string]ntopInterfaceFull) error {
	endpoint := fmt.Sprintf("%s%s%s?ifid=%d",
		c.config.Ntopng.EndPoint, luaRestV2Get, interfaceDataPath, interfaceId)
	req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setCommonOptions(req, false)

	body, status, _ := getHttpResponseBody(getHttpClient(c.config.Ntopng.AllowUnsafeTLS), req)
	if status != http.StatusOK {
		if body != nil {
			return fmt.Errorf("request to interface data endpoint was not successful. Status: '%d', Response: '%v'",
				status, *body)
		} else {
			return fmt.Errorf("request to interface data endpoint was not successful. Status: '%d'",
				status)
		}
	}

	rawInterface, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		return err
	}
	var ifFull ntopInterfaceFull
	err = json.Unmarshal(rawInterface, &ifFull)
	if err != nil {
		if ifName, err := c.ResolveIfID(interfaceId); err != nil {
			return fmt.Errorf("problem parsing ntop interface: %s - %v", ifName, err)
		} else {
			return fmt.Errorf("problem parsing ntop interface: %d - %v", interfaceId, err)
		}
	}
	tempInterfaces[ifFull.IfName] = ifFull
	return nil
}

func (c *Controller) ScrapeL7ProtocolEndpointForAllInterfaces() {
	tempL7Protocols := make(map[string]ntopL7Protocol)
	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		if err := c.scrapeL7ProtocolEndpoint(c.ifList[configuredIf], tempL7Protocols); err != nil {
			fmt.Printf("failed to scrape L7 protocols for interface '%s' with error: %v", configuredIf, err)
		}
	}
	c.ListRWMutex.Lock()
	defer c.ListRWMutex.Unlock()
	c.L7ProtocolList = tempL7Protocols
}

func (c *Controller) scrapeL7ProtocolEndpoint(interfaceId int, tempL7Protocols map[string]ntopL7Protocol) error {
	endpoint := fmt.Sprintf("%s%s%s?ifid=%d", c.config.Ntopng.EndPoint, luaRestV2Get, l7DataPath, interfaceId)
	req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setCommonOptions(req, false)

	body, status, _ := getHttpResponseBody(getHttpClient(c.config.Ntopng.AllowUnsafeTLS), req)
	if status != http.StatusOK {
		if body != nil {
			return fmt.Errorf("request to L7 protocol endpoint was not successful. Status: '%d', Response: '%v'",
				status, *body)
		}
		return fmt.Errorf("request to L7 protocol endpoint was not successful. Status: '%d'", status)
	}

	rawProtocols, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		return err
	}
	var protocols []ntopL7Protocol
	if err := json.Unmarshal(rawProtocols, &protocols); err != nil {
		return fmt.Errorf("was not able to parse L7 protocols from ntopng: %v", err)
	}
	ifName, err := c.ResolveIfID(interfaceId)
	if err != nil {
		return fmt.Errorf("could not resolve interface: %d: %v", interfaceId, err)
	}
	for _, protocol := range protocols {
		protocol.IfID = interfaceId
		protocol.IfName = ifName
		tempL7Protocols[fmt.Sprintf("%d:%s", interfaceId, protocol.Application.Name)] = protocol
	}
	return nil
}

func (c *Controller) setCommonOptions(req *http.Request, isJsonRequest bool) {
	if isJsonRequest {
		req.Header.Add("Content-Type", "application/json")
	}
	switch c.config.Ntopng.AuthMethod {
	case "cookie":
		req.Header.Add("Cookie",
			fmt.Sprintf("user=%s; password=%s",
				c.config.Ntopng.User, c.config.Ntopng.Password))
	case "basic":
		req.SetBasicAuth(c.config.Ntopng.User, c.config.Ntopng.Password)
	case "token":
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.config.Ntopng.Token))
	}
}

func getHttpClient(allowInsecure bool) *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	if allowInsecure {
		// #nosec G402 -- InsecureSkipVerify is intentionally configurable via AllowUnsafeTLS setting
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &http.Client{Transport: customTransport}
}
