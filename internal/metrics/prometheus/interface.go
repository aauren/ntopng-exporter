package prometheus

import (
	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	interfaceLabels  = []string{"ifname", "ifid"}
	tcpPacketLabels  = deepAppend(interfaceLabels, "type")
	throughputLabels = deepAppend(interfaceLabels, "direction")
)

type interfaceCollector struct {
	ntopNGController    *ntopng.Controller
	config              *config.Config
	alertedFlows        *prometheus.Desc
	alertedFlowsError   *prometheus.Desc
	alertedFlowsNotice  *prometheus.Desc
	alertedFlowsWarning *prometheus.Desc
	bytesRcvd           *prometheus.Desc
	bytesSent           *prometheus.Desc
	drops               *prometheus.Desc
	numDevices          *prometheus.Desc
	numHosts            *prometheus.Desc
	numLocalHosts       *prometheus.Desc
	packetsRcvd         *prometheus.Desc
	packetsSent         *prometheus.Desc
	speed               *prometheus.Desc
	tcpPacketStats      *prometheus.Desc
	throughputBPS       *prometheus.Desc
	throughputPPS       *prometheus.Desc
}

func NewNtopNGInterfaceCollector(ntopController *ntopng.Controller, config *config.Config) *interfaceCollector {
	return &interfaceCollector{
		ntopNGController: ntopController,
		config:           config,
		alertedFlows: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "alerted_flows"),
			"current number of alerted flows client flows",
			interfaceLabels,
			nil),
		alertedFlowsError: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "alerted_error_flows"),
			"current number of alerted error flows",
			interfaceLabels,
			nil),
		alertedFlowsNotice: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "alerted_notice_flows"),
			"current number of alerted notice flows",
			interfaceLabels,
			nil),
		alertedFlowsWarning: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "alerted_warning_flows"),
			"current number of alerted warning flows",
			interfaceLabels,
			nil),
		bytesRcvd: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "bytes_rcvd"),
			"total number of bytes received",
			interfaceLabels,
			nil),
		bytesSent: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "bytes_sent"),
			"total number of bytes sent",
			interfaceLabels,
			nil),
		drops: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "drops"),
			"number of drops",
			interfaceLabels,
			nil),
		numDevices: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "num_devices"),
			"number of devices",
			interfaceLabels,
			nil),
		numHosts: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "num_hosts"),
			"number of hosts",
			interfaceLabels,
			nil),
		numLocalHosts: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "num_local_hosts"),
			"number of hosts on the local network",
			interfaceLabels,
			nil),
		packetsRcvd: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "packets_rcvd"),
			"total number of packets received",
			interfaceLabels,
			nil),
		packetsSent: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "packets_sent"),
			"total number of packets sent",
			interfaceLabels,
			nil),
		speed: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "speed"),
			"current speed of interface in Mbps",
			interfaceLabels,
			nil),
		tcpPacketStats: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "tcp_packet_stats"),
			"tcp packet stats by type",
			tcpPacketLabels,
			nil),
		throughputBPS: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "current_throughput_bps"),
			"current throughput by direction in bytes per second",
			throughputLabels,
			nil),
		throughputPPS: prometheus.NewDesc(
			prometheus.BuildFQName("ntopng", "interface", "current_throughput_pps"),
			"current throughput by direction in packets per second",
			throughputLabels,
			nil),
	}
}

func (c *interfaceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.alertedFlows
	ch <- c.alertedFlowsError
	ch <- c.alertedFlowsNotice
	ch <- c.alertedFlowsWarning
	ch <- c.bytesRcvd
	ch <- c.bytesSent
	ch <- c.drops
	ch <- c.numDevices
	ch <- c.numHosts
	ch <- c.numLocalHosts
	ch <- c.packetsRcvd
	ch <- c.packetsSent
	ch <- c.speed
	ch <- c.tcpPacketStats
	ch <- c.throughputBPS
	ch <- c.throughputPPS
}

func (c *interfaceCollector) Collect(ch chan<- prometheus.Metric) {
	c.ntopNGController.ListRWMutex.RLock()
	defer c.ntopNGController.ListRWMutex.RUnlock()
	for _, myIf := range c.ntopNGController.InterfaceList {
		var interfaceLabelValues = []string{myIf.IfName, myIf.IfID}
		ch <- prometheus.MustNewConstMetric(c.alertedFlows, prometheus.GaugeValue, myIf.AlertedFlows,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.alertedFlowsError, prometheus.GaugeValue, myIf.AlertedFlowsError,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.alertedFlowsNotice, prometheus.GaugeValue, myIf.AlertedFlowsNotice,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.alertedFlowsWarning, prometheus.GaugeValue, myIf.AlertedFlowsWarning,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.bytesRcvd, prometheus.CounterValue, myIf.BytesReceived,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.bytesSent, prometheus.CounterValue, myIf.BytesSent,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.drops, prometheus.CounterValue, myIf.Drops,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.numDevices, prometheus.GaugeValue, myIf.NumDevices,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.numHosts, prometheus.GaugeValue, myIf.NumHosts,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.numLocalHosts, prometheus.GaugeValue, myIf.NumLocalHosts,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.packetsRcvd, prometheus.CounterValue, myIf.PacketsReceived,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.packetsSent, prometheus.CounterValue, myIf.PacketsSent,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.speed, prometheus.GaugeValue, myIf.Speed,
			interfaceLabelValues...)
		ch <- prometheus.MustNewConstMetric(c.tcpPacketStats, prometheus.CounterValue, myIf.TCPPacketStats.Lost,
			deepAppend(interfaceLabelValues, "lost")...)
		ch <- prometheus.MustNewConstMetric(c.tcpPacketStats, prometheus.CounterValue, myIf.TCPPacketStats.OutOfOrder,
			deepAppend(interfaceLabelValues, "out_of_order")...)
		ch <- prometheus.MustNewConstMetric(c.tcpPacketStats, prometheus.CounterValue, myIf.TCPPacketStats.Retransmissions,
			deepAppend(interfaceLabelValues, "retransmit")...)
		ch <- prometheus.MustNewConstMetric(c.throughputBPS, prometheus.GaugeValue, myIf.Throughput.Download.BPS,
			deepAppend(interfaceLabelValues, "received")...)
		ch <- prometheus.MustNewConstMetric(c.throughputBPS, prometheus.GaugeValue, myIf.Throughput.Upload.BPS,
			deepAppend(interfaceLabelValues, "sent")...)
		ch <- prometheus.MustNewConstMetric(c.throughputPPS, prometheus.GaugeValue, myIf.Throughput.Download.PPS,
			deepAppend(interfaceLabelValues, "received")...)
		ch <- prometheus.MustNewConstMetric(c.throughputPPS, prometheus.GaugeValue, myIf.Throughput.Download.PPS,
			deepAppend(interfaceLabelValues, "sent")...)
	}
}
