package prometheus

import (
	"strconv"

	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	hostLabels       = []string{"ip", "ifname", "mac", "name", "vlan"}
	basicDNSLabels   = deepAppend(hostLabels, "direction")
	DNSRepliesLabels = deepAppend(basicDNSLabels, "status")
	DNSQueriesLabels = deepAppend(basicDNSLabels, "record_type")
)

type hostCollector struct {
	ntopNGController  *ntopng.Controller
	config            *config.Config
	activeClientFlows *prometheus.Desc
	activeServerFlows *prometheus.Desc
	bytesRcvd         *prometheus.Desc
	bytesSent         *prometheus.Desc
	DNSQueryTypes     *prometheus.Desc
	numAlerts         *prometheus.Desc
	packetsRcvd       *prometheus.Desc
	packetsSent       *prometheus.Desc
	totalAlerts       *prometheus.Desc
	totalClientFlows  *prometheus.Desc
	totalDNSQueries   *prometheus.Desc
	totalDNSReplies   *prometheus.Desc
	totalServerFlows  *prometheus.Desc
}

func NewNtopNGHostCollector(ntopController *ntopng.Controller, config *config.Config) *hostCollector {
	return &hostCollector{
		ntopNGController: ntopController,
		config:           config,
		activeClientFlows: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "active_client_flows"),
			"current number of active client flows for host",
			hostLabels,
			nil),
		activeServerFlows: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "active_server_flows"),
			"current number of active server flows for host",
			hostLabels,
			nil),
		bytesRcvd: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "bytes_rcvd"),
			"number of bytes received for host",
			hostLabels,
			nil),
		bytesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "bytes_sent"),
			"number of bytes sent for host",
			hostLabels,
			nil),
		DNSQueryTypes: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "dns_queries_by_type"),
			"total number of DNS queries by record type",
			DNSQueriesLabels,
			nil),
		numAlerts: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "num_alerts"),
			"number of alerts for host",
			hostLabels,
			nil),
		packetsRcvd: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "packets_rcvd"),
			"number of packets received for host",
			hostLabels,
			nil),
		packetsSent: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "packets_sent"),
			"number of packets sent for host",
			hostLabels,
			nil),
		totalAlerts: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "total_alerts"),
			"total number of alerts for host",
			hostLabels,
			nil),
		totalClientFlows: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "total_client_flows"),
			"total number of client flows for host",
			hostLabels,
			nil),
		totalDNSQueries: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "total_dns_queries"),
			"total number of DNS queries for host",
			basicDNSLabels,
			nil),
		totalDNSReplies: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "total_dns_replies"),
			"total number of DNS replies for host by status",
			DNSRepliesLabels,
			nil),
		totalServerFlows: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "host", "total_server_flows"),
			"total number of server flows for host",
			hostLabels,
			nil),
	}
}

func (c *hostCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.bytesRcvd
	ch <- c.bytesSent
	ch <- c.activeClientFlows
	ch <- c.activeServerFlows
	ch <- c.totalDNSQueries
	ch <- c.totalDNSReplies
	ch <- c.DNSQueryTypes
	ch <- c.numAlerts
	ch <- c.totalAlerts
	ch <- c.totalClientFlows
	ch <- c.totalServerFlows
}

func (c *hostCollector) Collect(ch chan<- prometheus.Metric) {
	c.ntopNGController.ListRWMutex.RLock()
	defer c.ntopNGController.ListRWMutex.RUnlock()
	for _, host := range c.ntopNGController.HostList {
		var hostLabelValues = []string{host.IP, host.IfName, host.MAC, host.Name, strconv.Itoa(host.VLAN)}
		ch <- prometheus.MustNewConstMetric(c.activeClientFlows, prometheus.GaugeValue, host.ActiveFlowsAsClient,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.activeServerFlows, prometheus.GaugeValue, host.ActiveFlowsAsServer,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.bytesSent, prometheus.CounterValue, host.BytesSent, hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.bytesRcvd, prometheus.CounterValue, host.BytesReceived,
			hostLabelValues...)
		if !c.config.Metric.ExcludeDNSMetrics {
			c.outputDNSMetric(ch, "received", &host.DNS.Received, hostLabelValues)
			c.outputDNSMetric(ch, "sent", &host.DNS.Sent, hostLabelValues)
		}
		ch <- prometheus.MustNewConstMetric(c.numAlerts, prometheus.GaugeValue, host.NumAlerts,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.packetsRcvd, prometheus.CounterValue, host.PacketsReceived,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.packetsSent, prometheus.CounterValue, host.PacketsSent,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.totalAlerts, prometheus.CounterValue, host.TotalAlerts,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.totalClientFlows, prometheus.CounterValue, host.TotalFlowsAsClient,
			hostLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.totalServerFlows, prometheus.CounterValue, host.TotalFlowsAsServer,
			hostLabelValues...)
	}
}

func (c *hostCollector) outputDNSMetric(ch chan<- prometheus.Metric, direction string, dns *ntopng.NtopDNSSub,
	hostLabels []string) {
	dnsLabels := append(hostLabels, direction)
	ch <- prometheus.MustNewConstMetric(c.totalDNSQueries, prometheus.CounterValue, dns.NumQueries,
		dnsLabels...)
	ch <- prometheus.MustNewConstMetric(c.totalDNSReplies, prometheus.CounterValue, dns.NumRepliesError,
		deepAppend(dnsLabels, "error")...)
	ch <- prometheus.MustNewConstMetric(c.totalDNSReplies, prometheus.CounterValue, dns.NumRepliesOK,
		deepAppend(dnsLabels, "ok")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumA,
		deepAppend(dnsLabels, "A")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumAAAA,
		deepAppend(dnsLabels, "AAAA")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumAny,
		deepAppend(dnsLabels, "ANY")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumCName,
		deepAppend(dnsLabels, "CNAME")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumMX,
		deepAppend(dnsLabels, "MX")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumNS,
		deepAppend(dnsLabels, "NS")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumOther,
		deepAppend(dnsLabels, "OTHER")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumPTR,
		deepAppend(dnsLabels, "PTR")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumSOA,
		deepAppend(dnsLabels, "SOA")...)
	ch <- prometheus.MustNewConstMetric(c.DNSQueryTypes, prometheus.CounterValue, dns.Queries.NumTXT,
		deepAppend(dnsLabels, "TXT")...)
}
