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
	ActiveFlowsAsClient uint32  `json:"active_flows.as_client"`
	ActiveFlowsAsServer uint32  `json:"active_flows.as_server"`
	BytesReceived       uint64  `json:"bytes.rcvd"`
	BytesSent           uint64  `json:"bytes.sent"`
	DNS                 ntopDNS `json:"dns"`
	IfID                int     `json:"ifid"`
	IP                  string  `json:"IP"`
	MAC                 string  `json:"mac"`
	Name                string  `json:"name"`
	NumAlerts           uint32  `json:"num_alerts"`
	TotalAlerts         uint32  `json:"total_alerts"`
	TotalFlowsAsClient  uint32  `json:"total_flows.as_client"`
	TotalFlowsAsServer  uint32  `json:"total_flows.as_server"`
	VLAN                int     `json:"vlan"`
}

type ntopDNS struct {
	Received ntopDNSSub `json:"rcvd"`
	Sent     ntopDNSSub `json:"sent"`
}

type ntopDNSSub struct {
	NumQueries      uint32         `json:"num_queries"`
	NumRepliesError uint32         `json:"num_replies error"`
	NumRepliesOK    uint32         `json:"num_replies ok"`
	Queries         ntopDNSQueries `json:"queries"`
}

type ntopDNSQueries struct {
	NumA     uint32 `json:"num_a"`
	NumAAAA  uint32 `json:"num_aaaa"`
	NumAny   uint32 `json:"num_any"`
	NumCName uint32 `json:"num_cname"`
	NumMX    uint32 `json:"num_mx"`
	NumNS    uint32 `json:"num_ns"`
	NumOther uint32 `json:"num_other"`
	NumPTR   uint32 `json:"num_ptr"`
	NumSOA   uint32 `json:"num_soa"`
	NumTXT   uint32 `json:"num_txt"`
}

func (n ntopHost) String() string {
	output, _ := json.MarshalIndent(n, "", "\t")
	return fmt.Sprintf("%s", output)
}