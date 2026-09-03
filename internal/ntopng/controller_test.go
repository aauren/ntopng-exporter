package ntopng

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAuthMethod = "none"
	testHostIP     = "192.168.1.10"
	testIfName     = "eth0"
)

// TestHostKeyDistinguishesDimensions guards the cache key contract: IP, interface, and
// VLAN are independent dimensions, so the same IP seen on different interfaces or VLANs
// must never collapse into a single map entry. Dropping a field from hostKey breaks this.
func TestHostKeyDistinguishesDimensions(t *testing.T) {
	t.Parallel()
	base := hostKey{IP: testHostIP, IfID: 1, VLAN: 0}
	tests := []struct {
		name      string
		other     hostKey
		wantEqual bool
	}{
		{"identical keys collide", hostKey{IP: testHostIP, IfID: 1, VLAN: 0}, true},
		{"different ip is distinct", hostKey{IP: "10.0.0.1", IfID: 1, VLAN: 0}, false},
		{"different interface is distinct", hostKey{IP: testHostIP, IfID: 2, VLAN: 0}, false},
		{"different vlan is distinct", hostKey{IP: testHostIP, IfID: 1, VLAN: 20}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := map[hostKey]string{base: "first"}
			m[tt.other] = "second"
			if tt.wantEqual {
				assert.Len(t, m, 1, "identical keys should collapse to one entry")
			} else {
				assert.Len(t, m, 2, "distinct keys should remain separate entries")
			}
		})
	}
}

// ntopHostResponse mirrors the host custom_data endpoint envelope so we can drive
// scrapeHostEndpoint through a stub server using the real ntopHost JSON tags.
type ntopHostResponse struct {
	RcStr string     `json:"rc_str"`
	Rsp   []ntopHost `json:"rsp"`
}

// TestScrapeHostEndpointKeepsHostPerInterfaceAndVLAN is the regression guard for the
// composite key: ntopng reports the same IP under every interface and VLAN it's seen on,
// and every (ip, ifid, vlan) combination must survive instead of clobbering each other
// the way the old IP-only cache key did.
func TestScrapeHostEndpointKeepsHostPerInterfaceAndVLAN(t *testing.T) {
	t.Parallel()
	const (
		localIfID    = 1
		vlanIfID     = 2
		localIfName  = testIfName
		vlanIfName   = "eth1"
		localBytes   = 1000.0
		crossIfBytes = 50.0
		vlanBytes    = 75.0
		taggedVLAN   = 20
	)
	// Keyed by the ifid the stub is asked about, so each scrape returns that interface's view.
	responses := map[int][]ntopHost{
		localIfID: {
			{IP: testHostIP, IfID: localIfID, VLAN: 0, BytesSent: localBytes},
		},
		vlanIfID: {
			{IP: testHostIP, IfID: vlanIfID, VLAN: 0, BytesSent: crossIfBytes},
			{IP: testHostIP, IfID: vlanIfID, VLAN: taggedVLAN, BytesSent: vlanBytes},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
			return
		}
		var payload struct {
			IfID int `json:"ifid"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("unmarshaling request payload: %v", err)
			return
		}
		out, err := json.Marshal(ntopHostResponse{RcStr: "OK", Rsp: responses[payload.IfID]})
		if err != nil {
			t.Errorf("marshaling stub response: %v", err)
			return
		}
		if _, err := w.Write(out); err != nil {
			t.Errorf("writing stub response: %v", err)
		}
	}))
	defer server.Close()

	requestTimeout := 2 * time.Second
	c := CreateController(context.Background(), &config.Config{}, requestTimeout)
	assert.Equal(t, requestTimeout, c.httpClient.Timeout)
	c.config.Ntopng.EndPoint = server.URL
	c.config.Ntopng.AuthMethod = testAuthMethod
	c.ifList = map[string]int{localIfName: localIfID, vlanIfName: vlanIfID}

	hosts := make(map[hostKey]ntopHost)
	require.NoError(t, c.scrapeHostEndpoint(context.Background(), localIfID, hosts))
	require.NoError(t, c.scrapeHostEndpoint(context.Background(), vlanIfID, hosts))

	require.Len(t, hosts, 3, "every (ip, ifid, vlan) combination should be retained")

	checks := []struct {
		name      string
		key       hostKey
		wantBytes float64
		wantIf    string
	}{
		{"local interface retained", hostKey{IP: testHostIP, IfID: localIfID, VLAN: 0}, localBytes, localIfName},
		{"same ip on second interface retained", hostKey{IP: testHostIP, IfID: vlanIfID, VLAN: 0}, crossIfBytes, vlanIfName},
		{"same ip and interface on tagged vlan retained", hostKey{IP: testHostIP, IfID: vlanIfID, VLAN: taggedVLAN}, vlanBytes, vlanIfName},
	}
	for _, ck := range checks {
		got, ok := hosts[ck.key]
		require.Truef(t, ok, "%s: expected map entry to be present", ck.name)
		assert.Equalf(t, ck.wantBytes, got.BytesSent, "%s: bytes sent", ck.name)
		assert.Equalf(t, ck.wantIf, got.IfName, "%s: resolved interface name", ck.name)
	}
}

func TestScrapeL7ProtocolEndpoint(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/lua/rest/v2/get/interface/l7/data.lua", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("ifid"))
		_, err := w.Write([]byte(`{"rc_str":"OK","rsp":[{"application":{"name":"HTTP","id":7},` +
			`"bytes":{"rcvd":100,"sent":50,"total":150,"percentage":75},` +
			`"packets":{"rcvd":10,"sent":5,"total":15},"breed":"Safe","tot_num_flows":2}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	c := CreateController(context.Background(), &config.Config{}, config.DefaultRequestTimeout)
	c.config.Ntopng.EndPoint = server.URL
	c.config.Ntopng.AuthMethod = testAuthMethod
	c.ifList = map[string]int{testIfName: 1}
	protocols := make(map[string]ntopL7Protocol)

	require.NoError(t, c.scrapeL7ProtocolEndpoint(context.Background(), 1, protocols))
	protocol, ok := protocols["1:HTTP"]
	require.True(t, ok)
	assert.Equal(t, testIfName, protocol.IfName)
	assert.Equal(t, 100.0, protocol.Bytes.Received)
	assert.Equal(t, 15.0, protocol.Packets.Total)
	assert.Equal(t, 2.0, protocol.Flows)
}

func TestScrapeL7ProtocolEndpointHonorsControllerContext(t *testing.T) {
	t.Parallel()
	requestStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := CreateController(ctx, &config.Config{}, config.DefaultRequestTimeout)
	c.config.Ntopng.EndPoint = server.URL
	c.config.Ntopng.AuthMethod = testAuthMethod
	c.ifList = map[string]int{testIfName: 1}

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.scrapeL7ProtocolEndpoint(ctx, 1, make(map[string]ntopL7Protocol))
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the L7 request to start")
	}
	cancel()

	select {
	case err := <-errCh:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("L7 request did not stop after canceling the controller context")
	}
}

// TestScrapeCycleHonorsDeadline checks that a hung endpoint can't drag a scrape cycle past
// scrapeInterval, even when requestTimeout is much larger.
func TestScrapeCycleHonorsDeadline(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	myConfig := &config.Config{}
	myConfig.Ntopng.EndPoint = server.URL
	myConfig.Ntopng.AuthMethod = testAuthMethod
	myConfig.Ntopng.ScrapeInterval = "100ms"
	myConfig.Ntopng.ScrapeTargets = []string{config.L7Protocols}
	myConfig.Host.InterfacesToMonitor = []string{testIfName, "eth1"}

	c := CreateController(context.Background(), myConfig, config.DefaultRequestTimeout)
	c.ifList = map[string]int{testIfName: 1, "eth1": 2}

	done := make(chan struct{})
	go func() {
		c.ScrapeAllConfiguredTargets()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scrape cycle was not bounded by the per-cycle deadline")
	}
}
