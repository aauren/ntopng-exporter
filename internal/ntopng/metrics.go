package ntopng

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// interfaceListTarget labels the one request we make that isn't tied to an interface. Every other
// target value comes from config so a metric label matches what's under scrapeTargets.
const interfaceListTarget = "interfacelist"

// Reasons we break request failures down by on the errors counter. Anything we can't place lands in
// reasonRequest, which covers connection refused, DNS failures, TLS problems and the like
const (
	reasonTimeout  = "timeout"
	reasonDeadline = "deadline"
	reasonCanceled = "canceled"
	reasonStatus   = "http_status"
	reasonParse    = "parse"
	reasonEmpty    = "empty"
	reasonRequest  = "request"
)

// errUnexpectedStatus lets us tell a non-200 from ntopng apart from a transport failure without
// having to sniff error strings
var errUnexpectedStatus = errors.New("ntopng returned an unexpected status")

var (
	requestLabels      = []string{"ifname", "target"}
	requestErrorLabels = []string{"ifname", "target", "reason"}
	// Buckets stretch out to 60s because ntopng can take a very long time to answer on a busy
	// interface, and the client default tops out at 10s which would bunch every slow request into +Inf
	requestBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 60}
)

type requestMetrics struct {
	duration *prometheus.HistogramVec
	errors   *prometheus.CounterVec
}

func newRequestMetrics() *requestMetrics {
	return &requestMetrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    prometheus.BuildFQName("ntopng", "request", "duration_seconds"),
			Help:    "time ntopng took to answer a request, by interface and request type",
			Buckets: requestBuckets,
		}, requestLabels),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName("ntopng", "request", "errors_total"),
			Help: "requests to ntopng that timed out or otherwise failed, by interface, request type, and reason",
		}, requestErrorLabels),
	}
}

// observe records how long a request took and, when it failed, why. We time failures too so the
// histogram count reflects every request sent, timeouts included.
func (m *requestMetrics) observe(target, ifName string, duration time.Duration, err, reqCtxErr error) {
	m.duration.WithLabelValues(ifName, target).Observe(duration.Seconds())
	if err != nil {
		m.countError(target, ifName, classifyError(err, reqCtxErr))
	}
}

func (m *requestMetrics) countError(target, ifName, reason string) {
	m.errors.WithLabelValues(ifName, target, reason).Inc()
}

// classifyError sorts a failed request into requestTimeout, cycle deadline, shutdown, or connection
// reasons, checking the request's own context first since that's what separates a client Timeout from a canceled cycle.
func classifyError(err, reqCtxErr error) string {
	switch {
	case errors.Is(reqCtxErr, context.DeadlineExceeded):
		return reasonDeadline
	case errors.Is(reqCtxErr, context.Canceled), errors.Is(err, context.Canceled):
		return reasonCanceled
	case errors.Is(err, errUnexpectedStatus):
		return reasonStatus
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, os.ErrDeadlineExceeded):
		return reasonTimeout
	}
	// The net.Error fallback catches http.Client.Timeout, which older Go versions don't surface as a
	// context deadline
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return reasonTimeout
	}
	return reasonRequest
}

// Collectors exposes the request instrumentation so that main can register it alongside the other
// collectors, which keeps all of the registry wiring in one place
func (c *Controller) Collectors() []prometheus.Collector {
	return []prometheus.Collector{c.requestMetrics.duration, c.requestMetrics.errors}
}
