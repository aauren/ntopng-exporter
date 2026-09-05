package ntopng

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// timeoutError stands in for a transport level timeout so that we can exercise the net.Error branch
// of classifyError without waiting on a real one
type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestClassifyError(t *testing.T) {
	t.Parallel()
	var netErr net.Error = timeoutError{}
	// http.Client reports its own Timeout as a deadline error while leaving the request context
	// untouched, which is exactly how we tell it apart from the scrape cycle running out of interval
	tests := []struct {
		name      string
		err       error
		reqCtxErr error
		want      string
	}{
		{"non-200 from ntopng", fmt.Errorf("wrapped: %w", errUnexpectedStatus), nil, reasonStatus},
		{"request ran past requestTimeout", fmt.Errorf("wrapped: %w", context.DeadlineExceeded), nil, reasonTimeout},
		{"scrape cycle ran out of interval", fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
			context.DeadlineExceeded, reasonDeadline},
		{"socket deadline", fmt.Errorf("wrapped: %w", net.ErrClosed), nil, reasonRequest},
		{"shutdown mid-request", fmt.Errorf("wrapped: %w", context.Canceled), context.Canceled, reasonCanceled},
		{"shutdown reported by the client only", fmt.Errorf("wrapped: %w", context.Canceled), nil, reasonCanceled},
		{"transport timeout", fmt.Errorf("wrapped: %w", netErr), nil, reasonTimeout},
		{"anything else", errors.New("connection refused"), nil, reasonRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, classifyError(tt.err, tt.reqCtxErr))
		})
	}
}

// TestRequestMetricsRecordOutcomes checks that every request lands on the duration histogram under
// its own interface and target, and that failures land under the right reason on the errors counter.
func TestRequestMetricsRecordOutcomes(t *testing.T) {
	t.Parallel()
	const okBody = `{"rc_str":"OK","rsp":[]}`
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		closed     bool
		timeout    time.Duration
		wantErr    bool
		wantReason string
	}{
		{
			name: "successful request is timed and not counted as an error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(okBody))
			},
		},
		{
			name: "non-200 counts as an http status error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "boom", http.StatusInternalServerError)
			},
			wantErr:    true,
			wantReason: reasonStatus,
		},
		{
			name: "unparsable body counts as a parse error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("this is not json"))
			},
			wantErr:    true,
			wantReason: reasonParse,
		},
		{
			name: "hung endpoint counts as a timeout",
			handler: func(_ http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			},
			timeout:    50 * time.Millisecond,
			wantErr:    true,
			wantReason: reasonTimeout,
		},
		{
			name: "timeout while reading the body counts as a timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"rc_str":"OK","rsp":`))
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			},
			timeout:    50 * time.Millisecond,
			wantErr:    true,
			wantReason: reasonTimeout,
		},
		{
			name: "incomplete transfer of valid JSON counts as a request error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", fmt.Sprint(len(okBody)+1))
				_, _ = w.Write([]byte(okBody))
			},
			wantErr:    true,
			wantReason: reasonRequest,
		},
		{
			name:       "unreachable endpoint counts as a request error",
			handler:    func(_ http.ResponseWriter, _ *http.Request) {},
			closed:     true,
			wantErr:    true,
			wantReason: reasonRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tt.handler)
			if tt.closed {
				server.Close()
			} else {
				defer server.Close()
			}

			timeout := tt.timeout
			if timeout == 0 {
				timeout = config.DefaultRequestTimeout
			}
			c := CreateController(context.Background(), &config.Config{}, timeout)
			c.config.Ntopng.EndPoint = server.URL
			c.config.Ntopng.AuthMethod = testAuthMethod
			c.ifList = map[string]int{testIfName: 1}

			err := c.scrapeL7ProtocolEndpoint(context.Background(), 1, make(map[string]ntopL7Protocol))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, uint64(1), histogramCount(t, c.requestMetrics, testIfName, config.L7Protocols),
				"every request should be timed, whether or not it succeeded")
			if tt.wantReason == "" {
				assert.Zero(t, testutil.CollectAndCount(c.requestMetrics.errors),
					"a successful request should not create any error series")
				return
			}
			assert.Equal(t, 1, testutil.CollectAndCount(c.requestMetrics.errors),
				"a failed request should be counted under exactly one reason")
			assert.Equal(t, 1.0, testutil.ToFloat64(
				c.requestMetrics.errors.WithLabelValues(testIfName, config.L7Protocols, tt.wantReason)))
		})
	}
}

// TestRequestMetricsCountCycleDeadline checks that requests killed by the per-cycle deadline get
// their own reason and skip the duration histogram, since they were never actually sent.
func TestRequestMetricsCountCycleDeadline(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	const queuedIfName = "eth1"
	myConfig := &config.Config{}
	myConfig.Ntopng.EndPoint = server.URL
	myConfig.Ntopng.AuthMethod = testAuthMethod
	myConfig.Ntopng.ScrapeInterval = "100ms"
	myConfig.Ntopng.ScrapeTargets = []string{config.L7Protocols}
	myConfig.Ntopng.ParallelWorkers = 1
	myConfig.Host.InterfacesToMonitor = []string{testIfName, queuedIfName}

	// requestTimeout is far larger than the cycle deadline, so the deadline is what cuts these off
	c := CreateController(context.Background(), myConfig, 10*time.Second)
	c.ifList = map[string]int{testIfName: 1, queuedIfName: 2}
	c.ScrapeAllConfiguredTargets()

	for _, ifName := range []string{testIfName, queuedIfName} {
		assert.Equalf(t, 1.0, testutil.ToFloat64(
			c.requestMetrics.errors.WithLabelValues(ifName, config.L7Protocols, reasonDeadline)),
			"%s should be counted against the cycle deadline rather than requestTimeout", ifName)
	}
	assert.Equal(t, uint64(1), histogramCount(t, c.requestMetrics, testIfName, config.L7Protocols),
		"the request we actually sent should be timed")
	assert.Zero(t, histogramCount(t, c.requestMetrics, queuedIfName, config.L7Protocols),
		"a request abandoned before we sent it has no latency to report")
}

// TestRequestMetricsCountHostResponseErrors checks that we count each failed request only once.
func TestRequestMetricsCountHostResponseErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		payload    string
		wantReason string
		wantErr    bool
		wantHosts  int
	}{
		{"empty list", `[]`, reasonEmpty, true, 0},
		{"wrong payload type", `{}`, reasonParse, true, 0},
		{"partially decoded host", `[{"ip":"192.0.2.1","ifid":1,"bytes.sent":"invalid"}]`, reasonParse, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `{"rc_str":"OK","rsp":%s}`, tt.payload)
			}))
			defer server.Close()

			c := CreateController(context.Background(), &config.Config{}, config.DefaultRequestTimeout)
			c.config.Ntopng.EndPoint = server.URL
			c.config.Ntopng.AuthMethod = testAuthMethod
			c.ifList = map[string]int{testIfName: 1}

			hosts := make(map[hostKey]ntopHost)
			err := c.scrapeHostEndpoint(context.Background(), 1, hosts)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, hosts, tt.wantHosts)
			assert.Equal(t, uint64(1), histogramCount(t, c.requestMetrics, testIfName, config.HostScrape))
			assert.Equal(t, 1, testutil.CollectAndCount(c.requestMetrics.errors))
			assert.Equal(t, 1.0, testutil.ToFloat64(
				c.requestMetrics.errors.WithLabelValues(testIfName, config.HostScrape, tt.wantReason)))
		})
	}
}

// TestRequestMetricsLabelInterfaceList covers the one request we make that isn't tied to an
// interface, which we still want on the histogram so that a slow startup is visible
func TestRequestMetricsLabelInterfaceList(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := fmt.Fprintf(w, `{"rc_str":"OK","rsp":[{"ifname":%q,"ifid":1}]}`, testIfName)
		require.NoError(t, err)
	}))
	defer server.Close()

	myConfig := &config.Config{}
	myConfig.Ntopng.EndPoint = server.URL
	myConfig.Ntopng.AuthMethod = testAuthMethod
	c := CreateController(context.Background(), myConfig, config.DefaultRequestTimeout)

	require.NoError(t, c.CacheInterfaceIds())
	assert.Equal(t, uint64(1), histogramCount(t, c.requestMetrics, "", interfaceListTarget))
	assert.Zero(t, testutil.CollectAndCount(c.requestMetrics.errors))
}

// histogramCount pulls the observation count for one label pair straight back out of a registry,
// which is the only way to see how many requests we actually timed
func histogramCount(t *testing.T, metrics *requestMetrics, ifName, target string) uint64 {
	t.Helper()
	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metrics.duration))
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["ifname"] == ifName && labels["target"] == target {
				return metric.GetHistogram().GetSampleCount()
			}
		}
	}
	return 0
}
