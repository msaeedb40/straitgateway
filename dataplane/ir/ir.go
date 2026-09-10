// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package ir defines the Intermediate Representation (IR) types used as the
// strict boundary between the straitgateway control plane and the kernel dataplane.
//
// Architectural invariants:
//   - Controllers produce IR; they NEVER touch BPF maps directly.
//   - The Compiler is the ONLY component that translates IR → BPF/netlink state.
//   - Every IR object is generation-tracked for idempotent reconciliation.
package ir

import (
	"net/netip"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// Generation is a monotonically increasing revision counter.
// Used to detect and apply only delta changes to the dataplane.
type Generation int64

// ServiceIR is the intermediate representation of a Kubernetes Service.
// Produced by the Service controller; consumed by the Compiler.
type ServiceIR struct {
	// Generation is incremented on every change.
	Generation Generation
	// ID uniquely identifies this service.
	ID sgtypes.ServiceID
	// VIPv4 is the service ClusterIP (IPv4).
	VIPv4 netip.Addr
	// VIPv6 is the service ClusterIP (IPv6).
	VIPv6 netip.Addr
	// Port is the service port.
	Port uint16
	// Protocol is the L4 protocol.
	Protocol sgtypes.Protocol
	// Algorithm is the load balancing algorithm.
	Algorithm sgtypes.LBAlgorithm
	// Backends is the list of active backend endpoints.
	Backends []BackendIR
	// SessionAffinity enables sticky sessions (client IP hash).
	SessionAffinity bool
	// DSR enables Direct Server Return mode.
	DSR bool
	// IsNodePort indicates this service has a NodePort.
	IsNodePort bool
	// NodePort is the NodePort number (30000-32767).
	NodePort uint16
	// IsExternalLB indicates this service has an external LoadBalancer.
	IsExternalLB bool
}

// BackendIR is the IR for a single service backend endpoint.
type BackendIR struct {
	ID      sgtypes.BackendID
	Address netip.AddrPort
	NodeIP  netip.Addr
	Weight  uint32
	State   sgtypes.BackendState
}

// PolicyIR is the intermediate representation of a compiled network policy.
// Produced by the Policy controller from StraitNetworkPolicy + NetworkPolicy CRDs.
type PolicyIR struct {
	Generation     Generation
	ID             sgtypes.PolicyID
	SrcIdentity    sgtypes.Identity
	DstIdentity    sgtypes.Identity
	DstPort        uint16
	Protocol       sgtypes.Protocol
	Action         PolicyAction
	Priority       int32  // 0=highest, 255=lowest
	Direction      PolicyDirection
}

// PolicyAction defines the compiled policy action.
type PolicyAction uint8

const (
	PolicyActionAllow  PolicyAction = 0
	PolicyActionDeny   PolicyAction = 1
	PolicyActionReject PolicyAction = 2
)

// PolicyDirection defines the compiled policy direction.
type PolicyDirection uint8

const (
	PolicyDirectionIngress PolicyDirection = 0
	PolicyDirectionEgress  PolicyDirection = 1
)

// RouteIR is the intermediate representation of a FIB route.
// Produced by the Routing controller; compiled to netlink route operations.
type RouteIR struct {
	Generation Generation
	Dest       netip.Prefix
	Gateway    netip.Addr
	IfIndex    int
	Priority   int
	Table      int
	// ECMPNextHops lists ECMP next-hop gateways (empty = single-path).
	ECMPNextHops []netip.Addr
}

// NatIR is the intermediate representation of a NAT rule.
// Produced by the NAT controller; compiled to BPF ct_map entries.
type NatIR struct {
	Generation Generation
	// SNAT rules
	SrcCIDR   netip.Prefix
	SnatToIP  netip.Addr
	// DNAT rules
	DstIP     netip.Addr
	DstPort   uint16
	NatToIP   netip.Addr
	NatToPort uint16
	Protocol  sgtypes.Protocol
	Type      NatType
}

// NatType defines the type of NAT rule.
type NatType uint8

const (
	NatTypeSNAT       NatType = 0
	NatTypeDNAT       NatType = 1
	NatTypeMasquerade NatType = 2
	NatTypeNAT64      NatType = 3
)

// GatewayIR is the intermediate representation of a Gateway API gateway listener.
// Produced by the Gateway controller; compiled to eBPF service maps.
type GatewayIR struct {
	Generation  Generation
	ID          sgtypes.GatewayID
	GatewayName string
	Namespace   string
	// Listeners is the list of gateway listeners.
	Listeners []ListenerIR
}

// ListenerIR is the IR for a single Gateway API listener.
type ListenerIR struct {
	Name     string
	Protocol string // HTTP, HTTPS, TCP, UDP, TLS
	Port     uint16
	// Routes maps route names to their compiled route rules.
	Routes []RouteRuleIR
}

// RouteRuleIR is the compiled form of a Gateway API route rule.
type RouteRuleIR struct {
	// Backends is the list of weighted backend references.
	Backends []BackendIR
	// PathPrefix is the HTTP path prefix (empty for non-HTTP routes).
	PathPrefix string
	// Headers are HTTP header match conditions.
	Headers map[string]string
}

// TransitIR is the intermediate representation of transit gateway state.
// Produced by the Transit controller; compiled to WireGuard/IPsec tunnel config.
type TransitIR struct {
	Generation Generation
	SegmentID  sgtypes.SegmentID
	LocalIP    netip.Addr
	Peers      []TransitPeerIR
	Routes     []RouteIR
}

// TransitPeerIR is the IR for a remote cluster transit peer.
type TransitPeerIR struct {
	ClusterID    sgtypes.ClusterID
	TunnelIP     netip.Addr
	WGPublicKey  [32]byte
	AllowedCIDRs []netip.Prefix
	SegmentID    sgtypes.SegmentID
}

// IdentityIR is the IR for a security identity assignment.
// Produced by the Identity controller; compiled to identity_map entries.
type IdentityIR struct {
	Generation  Generation
	Identity    sgtypes.Identity
	LabelsHash  uint64
	SegmentID   sgtypes.SegmentID
	Namespace   string
	PodName     string
}

// DataplaneState is the complete desired state produced by merging all IR sources.
// This is what the Compiler consumes to produce the final BPF map state.
type DataplaneState struct {
	Generation  Generation
	Services    []ServiceIR
	Policies    []PolicyIR
	Routes      []RouteIR
	NatRules    []NatIR
	Gateways    []GatewayIR
	Transit     []TransitIR
	Identities  []IdentityIR
}
