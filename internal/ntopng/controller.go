package ntopng

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

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
	ctx            context.Context
	httpClient     *http.Client
	ifList         map[string]int
	requestMetrics *requestMetrics
	HostList       map[hostKey]ntopHost
	InterfaceList  map[string]ntopInterfaceFull
	L7ProtocolList map[string]ntopL7Protocol
	ListRWMutex    *sync.RWMutex
}

func CreateController(ctx context.Context, myConfig *config.Config, requestTimeout time.Duration) Controller {
	var controller Controller
	controller.config = myConfig
	controller.ctx = ctx
	controller.httpClient = getHttpClient(myConfig.Ntopng.AllowUnsafeTLS, requestTimeout)
	controller.requestMetrics = newRequestMetrics()
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
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("scrape interval hit: scraping from ntop\n")
			c.ScrapeAllConfiguredTargets()
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Controller) ScrapeAllConfiguredTargets() {
	// requestTimeout alone can't stop a cycle from overrunning the interval, so we cap the whole cycle
	// too. ParseConfig already validated it, so we skip the deadline if parsing somehow fails here.
	cycleCtx := c.ctx
	if scrapeInterval, err := time.ParseDuration(c.config.Ntopng.ScrapeInterval); err == nil && scrapeInterval > 0 {
		var cancel context.CancelFunc
		cycleCtx, cancel = context.WithTimeout(c.ctx, scrapeInterval)
		defer cancel()
	}

	// All enabled targets share one worker pool so parallelWorkers caps concurrency globally. Results
	// land in temp maps first, replaced wholesale, to minimize lock time and avoid unbounded growth.
	group := c.newScrapeGroup()
	scrapeHosts := c.isTargetEnabled(config.HostScrape)
	scrapeInterfaces := c.isTargetEnabled(config.InterfaceScrape)
	scrapeL7 := c.isTargetEnabled(config.L7Protocols)
	var tempNtopHosts map[hostKey]ntopHost
	var tempNtopInterfaces map[string]ntopInterfaceFull
	var tempL7Protocols map[string]ntopL7Protocol
	if scrapeHosts {
		tempNtopHosts = queueScrapes(cycleCtx, c, group, "hosts", c.scrapeHostEndpoint)
	}
	if scrapeInterfaces {
		tempNtopInterfaces = queueScrapes(cycleCtx, c, group, "interface data", c.scrapeInterfaceEndpoint)
	}
	if scrapeL7 {
		tempL7Protocols = queueScrapes(cycleCtx, c, group, "L7 protocols", c.scrapeL7ProtocolEndpoint)
	}
	_ = group.Wait()

	c.ListRWMutex.Lock()
	defer c.ListRWMutex.Unlock()
	if scrapeHosts {
		c.HostList = tempNtopHosts
	}
	if scrapeInterfaces {
		c.InterfaceList = tempNtopInterfaces
	}
	if scrapeL7 {
		c.L7ProtocolList = tempL7Protocols
	}
}

func (c *Controller) isTargetEnabled(target string) bool {
	return internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, target) ||
		internal.IsItemInArray(c.config.Ntopng.ScrapeTargets, config.AllScrape)
}

func (c *Controller) CacheInterfaceIds() error {
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV2Get, interfaceListPath)
	req, err := http.NewRequestWithContext(c.ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to get response from ntopng interface endpoint: %v", err)
	}
	c.setCommonOptions(req, false)

	// We have no interface name to label with yet, which is the whole point of this call, so the
	// ifname label stays empty for it
	rawInterfaces, err := c.getNtopData(req, interfaceListTarget, "")
	if err != nil {
		return err
	}
	var ifList []ntopInterface
	if err = json.Unmarshal(rawInterfaces, &ifList); err != nil {
		c.requestMetrics.countError(interfaceListTarget, "", reasonParse)
		return fmt.Errorf("was not able to parse interface list from ntopng: %v", err)
	}
	if len(ifList) < 1 {
		c.requestMetrics.countError(interfaceListTarget, "", reasonEmpty)
		return fmt.Errorf("ntopng returned 0 interfaces")
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

// newScrapeGroup builds the shared worker pool for a scrape cycle, clamped to 1 so a zero-value
// config can't deadlock SetLimit (validate() already rejects anything below 1 from a config file).
func (c *Controller) newScrapeGroup() *errgroup.Group {
	var group errgroup.Group
	group.SetLimit(max(c.config.Ntopng.ParallelWorkers, 1))
	return &group
}

// queueScrapes submits one scrape per monitored interface onto the shared worker group. Each worker
// fills a local map so scrapes don't need their own locking; the merged map is only safe to read after group.Wait().
func queueScrapes[K comparable, V any](ctx context.Context, c *Controller, group *errgroup.Group, target string,
	scrape func(ctx context.Context, interfaceId int, results map[K]V) error) map[K]V {
	merged := make(map[K]V)
	mergeMutex := &sync.Mutex{}
	for _, configuredIf := range c.config.Host.InterfacesToMonitor {
		group.Go(func() error {
			local := make(map[K]V)
			if err := scrape(ctx, c.ifList[configuredIf], local); err != nil {
				// We stay quiet on context.Canceled because it just means we're shutting down mid-scrape,
				// and we always return nil so that one bad interface doesn't stop the others
				if !errors.Is(err, context.Canceled) {
					fmt.Printf("failed to scrape %s for interface '%s' with error: %v", target, configuredIf, err)
				}
				return nil
			}
			mergeMutex.Lock()
			defer mergeMutex.Unlock()
			maps.Copy(merged, local)
			return nil
		})
	}
	return merged
}

func (c *Controller) scrapeHostEndpoint(ctx context.Context, interfaceId int, tempNtopHosts map[hostKey]ntopHost) error {
	ifName := c.resolveIfNameOrID(interfaceId)
	endpoint := fmt.Sprintf("%s%s%s", c.config.Ntopng.EndPoint, luaRestV2Get, hostCustomPath)
	payload := []byte(fmt.Sprintf(`{"ifid": %d, "field_alias": "%s"}`, interfaceId, hostCustomFields))
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	c.setCommonOptions(req, true)

	rawHosts, err := c.getNtopData(req, config.HostScrape, ifName)
	if err != nil {
		return err
	}
	var hostList []ntopHost
	if err = json.Unmarshal(rawHosts, &hostList); err != nil {
		// We'd rather keep whatever hosts did decode than throw the whole interface away over one bad
		// field, so we only count this instead of bailing out the way the other endpoints do
		c.requestMetrics.countError(config.HostScrape, ifName, reasonParse)
		if len(hostList) == 0 {
			return fmt.Errorf("was not able to parse hosts from ntopng: %w", err)
		}
	}
	if len(hostList) < 1 {
		c.requestMetrics.countError(config.HostScrape, ifName, reasonEmpty)
		return fmt.Errorf("ntopng returned 0 hosts")
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

func (c *Controller) scrapeInterfaceEndpoint(ctx context.Context, interfaceId int, tempInterfaces map[string]ntopInterfaceFull) error {
	ifName := c.resolveIfNameOrID(interfaceId)
	endpoint := fmt.Sprintf("%s%s%s?ifid=%d",
		c.config.Ntopng.EndPoint, luaRestV2Get, interfaceDataPath, interfaceId)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setCommonOptions(req, false)

	rawInterface, err := c.getNtopData(req, config.InterfaceScrape, ifName)
	if err != nil {
		return err
	}
	var ifFull ntopInterfaceFull
	if err = json.Unmarshal(rawInterface, &ifFull); err != nil {
		c.requestMetrics.countError(config.InterfaceScrape, ifName, reasonParse)
		return fmt.Errorf("problem parsing ntop interface: %s - %v", ifName, err)
	}
	tempInterfaces[ifFull.IfName] = ifFull
	return nil
}

func (c *Controller) scrapeL7ProtocolEndpoint(ctx context.Context, interfaceId int, tempL7Protocols map[string]ntopL7Protocol) error {
	ifName := c.resolveIfNameOrID(interfaceId)
	endpoint := fmt.Sprintf("%s%s%s?ifid=%d", c.config.Ntopng.EndPoint, luaRestV2Get, l7DataPath, interfaceId)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	c.setCommonOptions(req, false)

	rawProtocols, err := c.getNtopData(req, config.L7Protocols, ifName)
	if err != nil {
		return err
	}
	var protocols []ntopL7Protocol
	if err := json.Unmarshal(rawProtocols, &protocols); err != nil {
		c.requestMetrics.countError(config.L7Protocols, ifName, reasonParse)
		return fmt.Errorf("was not able to parse L7 protocols from ntopng: %v", err)
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

// getNtopData is the single choke point every ntopng API call goes through for timing, error
// counting, and envelope unwrapping. It only returns a payload on a 200 OK envelope.
func (c *Controller) getNtopData(req *http.Request, target, ifName string) (json.RawMessage, error) {
	// A scrape queued behind a slow one can find the cycle already over before a worker frees up. We
	// count that but skip sending, so a request that never left the machine doesn't skew the latency histogram.
	if ctxErr := req.Context().Err(); ctxErr != nil {
		c.requestMetrics.countError(target, ifName, classifyError(ctxErr, ctxErr))
		return nil, fmt.Errorf("abandoned %s request before sending it: %w", target, ctxErr)
	}

	start := time.Now()
	body, status, err := getHttpResponseBody(c.httpClient, req)
	if err == nil && status != http.StatusOK {
		err = fmt.Errorf("%w: '%d'%s", errUnexpectedStatus, status, responseSuffix(body))
	}
	c.requestMetrics.observe(target, ifName, time.Since(start), err, req.Context().Err())
	if err != nil {
		return nil, fmt.Errorf("request to %s endpoint: %w", target, err)
	}

	raw, err := getRawJsonFromNtopResponse(body)
	if err != nil {
		c.requestMetrics.countError(target, ifName, reasonParse)
		return nil, fmt.Errorf("failed to parse %s response from ntopng: %v%s", target, err, responseSuffix(body))
	}
	return raw, nil
}

func getHttpClient(allowInsecure bool, requestTimeout time.Duration) *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	if allowInsecure {
		// #nosec G402 -- InsecureSkipVerify is intentionally configurable via AllowUnsafeTLS setting
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &http.Client{Transport: customTransport, Timeout: requestTimeout}
}
