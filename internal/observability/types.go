package observability

import "time"

// DataplaneMetrics represents high-level metrics exported to Prometheus and the UI
type DataplaneMetrics struct {
	PacketsTotal        uint64    `json:"packetsTotal"`
	BytesTotal          uint64    `json:"bytesTotal"`
	DropsTotal          uint64    `json:"dropsTotal"`
	PolicyDeniedTotal   uint64    `json:"policyDeniedTotal"`
	RouteLookupTotal    uint64    `json:"routeLookupTotal"`
	RouteMissTotal      uint64    `json:"routeMissTotal"`
	ConntrackEntries    uint32    `json:"conntrackEntries"`
	NATTranslations     uint32    `json:"natTranslations"`
	BGPSessionsActive   int       `json:"bgpSessionsActive"`
	BFDSessionsActive   int       `json:"bfdSessionsActive"`
	WireGuardPeersCount int       `json:"wireguardPeersCount"`
	WireGuardRxBytes    uint64    `json:"wireguardRxBytes"`
	WireGuardTxBytes    uint64    `json:"wireguardTxBytes"`
	ServiceLBRequests   uint64    `json:"serviceLbRequests"`
	Timestamp           time.Time `json:"timestamp"`
}

// FlowLogEntry models an individual packet/flow telemetry event
type FlowLogEntry struct {
	ID                  string    `json:"id"`
	Timestamp           time.Time `json:"timestamp"`
	SourceIP            string    `json:"sourceIp"`
	DestinationIP       string    `json:"destinationIp"`
	SourcePort          uint16    `json:"sourcePort"`
	DestinationPort     uint16    `json:"destinationPort"`
	Protocol            string    `json:"protocol"` // TCP, UDP, ICMP
	SourceIdentity      uint32    `json:"sourceIdentity"`
	DestinationIdentity uint32    `json:"destinationIdentity"`
	Action              string    `json:"action"` // ALLOW, DENY, REDIRECT
	Node                string    `json:"node"`
	Interface           string    `json:"interface"`
	Bytes               uint64    `json:"bytes"`
	Packets             uint64    `json:"packets"`
	RuleName            string    `json:"ruleName,omitempty"`
}

// EbpfProgramStatus represents kernel verification and attachment state of eBPF programs
type EbpfProgramStatus struct {
	ID             uint32 `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"` // XDP, TCX_INGRESS, TCX_EGRESS, CGROUP_SOCK_ADDR, NETKIT
	Device         string `json:"device"`
	Instructions   uint32 `json:"instructions"`
	RunTimeTotalNs uint64 `json:"runTimeTotalNs"`
	RunCount       uint64 `json:"runCount"`
	Verified       bool   `json:"verified"`
}
