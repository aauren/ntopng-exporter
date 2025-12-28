package ntopng

import (
	"encoding/json"
)

type ntopResponse struct {
	RcStr string `json:"rc_str"`
	Rc    int
	Rsp   json.RawMessage
}

type ntopInterface struct {
	IfID   int    `json:"ifid"`
	IfName string `json:"ifname"`
}

type ntopHost struct {
	ActiveFlowsAsClient float64 `json:"active_flows.as_client"`
	ActiveFlowsAsServer float64 `json:"active_flows.as_server"`
	BytesReceived       float64 `json:"bytes.rcvd"`
	BytesSent           float64 `json:"bytes.sent"`
	DNS                 ntopDNS `json:"dns"`
	IfID                int     `json:"ifid"`
	IfName              string  `json:"ifname"`
	IP                  string  `json:"IP"`
	MAC                 string  `json:"mac"`
	Name                string  `json:"name"`
	NumAlerts           float64 `json:"num_alerts"`
	PacketsReceived     float64 `json:"packets.rcvd"`
	PacketsSent         float64 `json:"packets.sent"`
	TotalAlerts         float64 `json:"total_alerts"`
	TotalFlowsAsClient  float64 `json:"total_flows.as_client"`
	TotalFlowsAsServer  float64 `json:"total_flows.as_server"`
	VLAN                int     `json:"vlan"`
}

type ntopDNS struct {
	Received NtopDNSSub `json:"rcvd"`
	Sent     NtopDNSSub `json:"sent"`
}

type NtopDNSSub struct {
	NumQueries      float64        `json:"num_queries"`
	NumRepliesError float64        `json:"num_replies error"`
	NumRepliesOK    float64        `json:"num_replies ok"`
	Queries         ntopDNSQueries `json:"queries"`
}

type ntopDNSQueries struct {
	NumA     float64 `json:"num_a"`
	NumAAAA  float64 `json:"num_aaaa"`
	NumAny   float64 `json:"num_any"`
	NumCName float64 `json:"num_cname"`
	NumMX    float64 `json:"num_mx"`
	NumNS    float64 `json:"num_ns"`
	NumOther float64 `json:"num_other"`
	NumPTR   float64 `json:"num_ptr"`
	NumSOA   float64 `json:"num_soa"`
	NumTXT   float64 `json:"num_txt"`
}

type ntopInterfaceFull struct {
	AlertedFlows        float64            `json:"alerted_flows"`
	AlertedFlowsError   float64            `json:"alerted_flows_error"`
	AlertedFlowsNotice  float64            `json:"alerted_flows_notice"`
	AlertedFlowsWarning float64            `json:"alerted_flows_warning"`
	BytesReceived       float64            `json:"bytes_download"`
	BytesSent           float64            `json:"bytes_upload"`
	Drops               float64            `json:"drops"`
	IfID                string             `json:"ifid"`
	IfName              string             `json:"ifname"`
	NumDevices          float64            `json:"num_devices"`
	NumHosts            float64            `json:"num_hosts"`
	NumLocalHosts       float64            `json:"num_local_hosts"`
	PacketsReceived     float64            `json:"packets_download"`
	PacketsSent         float64            `json:"packets_upload"`
	Speed               float64            `json:"speed"`
	TCPPacketStats      ntopTCPPacketStats `json:"tcpPacketStats"`
	Throughput          ntopThroughput     `json:"throughput"`
}

type ntopTCPPacketStats struct {
	Lost            float64 `json:"lost"`
	OutOfOrder      float64 `json:"out_of_order"`
	Retransmissions float64 `json:"retransmissions"`
}

type ntopThroughput struct {
	Download ntopThroughputSub `json:"download"`
	Upload   ntopThroughputSub `json:"upload"`
}

type ntopThroughputSub struct {
	BPS float64 `json:"bps"`
	PPS float64 `json:"pps"`
}

func (n ntopHost) String() string {
	output, _ := json.MarshalIndent(n, "", "\t")
	return string(output)
}

func (n ntopInterfaceFull) String() string {
	output, _ := json.MarshalIndent(n, "", "\t")
	return string(output)
}
