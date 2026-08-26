package prometheus

import (
	"strconv"

	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
)

var l7Labels = []string{"ifname", "ifid", "application", "application_id", "breed"}

type l7Collector struct {
	ntopNGController *ntopng.Controller
	bytesReceived    *prometheus.Desc
	bytesSent        *prometheus.Desc
	bytesTotal       *prometheus.Desc
	packetsReceived  *prometheus.Desc
	packetsSent      *prometheus.Desc
	packetsTotal     *prometheus.Desc
	flows            *prometheus.Desc
}

func NewNtopNGL7Collector(ntopController *ntopng.Controller, _ *config.Config) *l7Collector {
	return &l7Collector{
		ntopNGController: ntopController,
		bytesReceived: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "bytes_rcvd"),
			"total number of bytes received by application protocol", l7Labels, nil),
		bytesSent: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "bytes_sent"),
			"total number of bytes sent by application protocol", l7Labels, nil),
		bytesTotal: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "bytes_total"),
			"total number of bytes by application protocol", l7Labels, nil),
		packetsReceived: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "packets_rcvd"),
			"total number of packets received by application protocol", l7Labels, nil),
		packetsSent: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "packets_sent"),
			"total number of packets sent by application protocol", l7Labels, nil),
		packetsTotal: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "packets_total"),
			"total number of packets by application protocol", l7Labels, nil),
		flows: prometheus.NewDesc(prometheus.BuildFQName("ntopng", "l7", "flows"),
			"total number of flows by application protocol", l7Labels, nil),
	}
}

func (c *l7Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.bytesReceived
	ch <- c.bytesSent
	ch <- c.bytesTotal
	ch <- c.packetsReceived
	ch <- c.packetsSent
	ch <- c.packetsTotal
	ch <- c.flows
}

func (c *l7Collector) Collect(ch chan<- prometheus.Metric) {
	c.ntopNGController.ListRWMutex.RLock()
	defer c.ntopNGController.ListRWMutex.RUnlock()
	for _, protocol := range c.ntopNGController.L7ProtocolList {
		labels := []string{protocol.IfName, strconv.Itoa(protocol.IfID), protocol.Application.Name,
			strconv.Itoa(protocol.Application.ID), protocol.Breed}
		ch <- prometheus.MustNewConstMetric(c.bytesReceived, prometheus.CounterValue, protocol.Bytes.Received, labels...)
		ch <- prometheus.MustNewConstMetric(c.bytesSent, prometheus.CounterValue, protocol.Bytes.Sent, labels...)
		ch <- prometheus.MustNewConstMetric(c.bytesTotal, prometheus.CounterValue, protocol.Bytes.Total, labels...)
		ch <- prometheus.MustNewConstMetric(c.packetsReceived, prometheus.CounterValue, protocol.Packets.Received, labels...)
		ch <- prometheus.MustNewConstMetric(c.packetsSent, prometheus.CounterValue, protocol.Packets.Sent, labels...)
		ch <- prometheus.MustNewConstMetric(c.packetsTotal, prometheus.CounterValue, protocol.Packets.Total, labels...)
		ch <- prometheus.MustNewConstMetric(c.flows, prometheus.GaugeValue, protocol.Flows, labels...)
	}
}
