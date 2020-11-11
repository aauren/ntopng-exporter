package ntopng

import (
	"encoding/json"
	"fmt"
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
	ActiveFlowsAsClient float64  `json:"active_flows.as_client"`
	ActiveFlowsAsServer float64  `json:"active_flows.as_server"`
	BytesReceived       float64  `json:"bytes.rcvd"`
	BytesSent           float64  `json:"bytes.sent"`
	DNS                 ntopDNS `json:"dns"`
	IfID                int     `json:"ifid"`
	IfName 				string	`json:ifname`
	IP                  string  `json:"IP"`
	MAC                 string  `json:"mac"`
	Name                string  `json:"name"`
	NumAlerts           float64  `json:"num_alerts"`
	TotalAlerts         float64  `json:"total_alerts"`
	TotalFlowsAsClient  float64  `json:"total_flows.as_client"`
	TotalFlowsAsServer  float64  `json:"total_flows.as_server"`
	VLAN                int     `json:"vlan"`
}

type ntopDNS struct {
	Received NtopDNSSub `json:"rcvd"`
	Sent     NtopDNSSub `json:"sent"`
}

type NtopDNSSub struct {
	NumQueries      float64         `json:"num_queries"`
	NumRepliesError float64         `json:"num_replies error"`
	NumRepliesOK    float64         `json:"num_replies ok"`
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

func (n ntopHost) String() string {
	output, _ := json.MarshalIndent(n, "", "\t")
	return fmt.Sprintf("%s", output)
}