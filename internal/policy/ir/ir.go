// Package ir defines the canonical Intermediate Representation for StraitGateway policies.
package ir

// Action specifies whether matching traffic is allowed or dropped.
type Action string

const (
	ActionAllow Action = "ALLOW"
	ActionDrop  Action = "DROP"
)

// Direction defines traffic flow relative to the workload.
type Direction string

const (
	DirectionIngress Direction = "INGRESS"
	DirectionEgress  Direction = "EGRESS"
)

// Protocol defines network protocol for filtering.
type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
	ProtocolICMP Protocol = "ICMP"
	ProtocolAny  Protocol = "ANY"
)

// PortRange specifies a port or range of ports.
type PortRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
}

// Selector matches workload identity, namespaces, or labels.
type Selector struct {
	Namespaces []string          `json:"namespaces,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CIDRs      []string          `json:"cidrs,omitempty"`
}

// Rule is a single compiled policy rule in Intermediate Representation.
type Rule struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Direction   Direction   `json:"direction"`
	Action      Action      `json:"action"`
	Protocol    Protocol    `json:"protocol"`
	Ports       []PortRange `json:"ports,omitempty"`
	Source      Selector    `json:"source"`
	Destination Selector    `json:"destination"`
	CELExpr     string      `json:"celExpr,omitempty"`
	Priority    int32       `json:"priority"`
}

// PolicySet is a collection of compiled policy rules applied to a workload.
type PolicySet struct {
	WorkloadID string `json:"workloadId"`
	Rules      []Rule `json:"rules"`
}
