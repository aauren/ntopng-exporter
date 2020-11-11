package metrics

import (
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

var (
	hostLabels = []string{"ip", "ifname", "mac", "name", "vlan"}
)

type ntopNGCollector struct {
	ntopNGController *ntopng.Controller
	bytesRcvd        *prometheus.Desc
	bytesSent        *prometheus.Desc
}

func NewNtopNGCollector(ntopController *ntopng.Controller) *ntopNGCollector {
	return &ntopNGCollector{
		ntopNGController: ntopController,
		bytesRcvd: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "", "bytes_rcvd"),
			"Number of bytes received by host",
			hostLabels,
			nil),
		bytesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "", "bytes_sent"),
			"Number of bytes sent by host",
			hostLabels,
			nil),
	}
}

func (c *ntopNGCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.bytesRcvd
	ch <- c.bytesSent
}

func (c *ntopNGCollector) Collect(ch chan<- prometheus.Metric) {
	for _, host := range c.ntopNGController.HostList {
		var hostLabelValues = []string{host.IP, host.IfName, host.MAC, host.Name, strconv.Itoa(host.VLAN)}
		ch <- prometheus.MustNewConstMetric(c.bytesSent, prometheus.CounterValue, host.BytesSent, hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.bytesRcvd, prometheus.CounterValue, host.BytesReceived, hostLabelValues...)
	}
}
