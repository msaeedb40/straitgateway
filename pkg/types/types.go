// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package types defines core types shared across the straitgateway codebase.
package types

import "net/netip"

// Identity is a 32-bit network identity derived from label selectors
// (pod, namespace, cluster, segment). Used as BPF map keys.
type Identity uint32

const (
	// IdentityUnknown represents an unresolved identity.
	IdentityUnknown Identity = 0
	// IdentityWorld represents traffic from outside the cluster.
	IdentityWorld Identity = 1
	// IdentityHost represents traffic from the host network namespace.
	IdentityHost Identity = 2
	// IdentityInit represents pods that have not yet received an identity.
	IdentityInit Identity = 3
	// IdentityReservedMin is the start of reserved identity range.
	IdentityReservedMin Identity = 0
	// IdentityReservedMax is the end of reserved identity range.
	IdentityReservedMax Identity = 255
	// IdentityLocalMin is the start of locally-allocated identities.
	IdentityLocalMin Identity = 256
)

// SegmentID is a 32-bit transit segment identifier.
// Segment 0 is the backbone segment; all others are isolated by default.
type SegmentID uint32

const (
	// SegmentBackbone is the default backbone segment.
	SegmentBackbone SegmentID = 0
)

// ClusterID is a unique cluster identifier in the transit mesh.
type ClusterID uint32

// ServiceID uniquely identifies a Kubernetes Service within a cluster.
type ServiceID struct {
	ClusterID ClusterID
	Namespace string
	Name      string
}

// BackendID uniquely identifies a service backend endpoint.
type BackendID uint32

// PolicyID uniquely identifies a compiled network policy rule.
type PolicyID uint32

// GatewayID uniquely identifies a Gateway API Gateway.
type GatewayID uint32

// FlowID is a unique per-flow identifier for observability correlation.
type FlowID uint64

// TraceID is a W3C Trace Context compatible trace identifier.
type TraceID [16]byte

// Endpoint represents a pod network endpoint.
type Endpoint struct {
	// ID is the local endpoint identifier (used as BPF map key).
	ID uint32
	// Identity is the security identity of this endpoint.
	Identity Identity
	// IPv4 is the pod's IPv4 address.
	IPv4 netip.Addr
	// IPv6 is the pod's IPv6 address.
	IPv6 netip.Addr
	// NodeIP is the host node's IP address.
	NodeIP netip.Addr
	// Namespace is the Kubernetes namespace.
	Namespace string
	// PodName is the Kubernetes pod name.
	PodName string
	// NetNS is the Linux network namespace path.
	NetNS string
	// IfIndex is the host-side NetKit interface index.
	IfIndex int
	// ContainerIfIndex is the container-side NetKit interface index.
	ContainerIfIndex int
	// MAC is the MAC address of the host-side NetKit interface.
	MAC [6]byte
}

// LBAlgorithm defines the load balancing algorithm for a service.
// +kubebuilder:validation:Enum=Maglev;RoundRobin;LeastConnections;IPHash;Random
type LBAlgorithm string

const (
	LBAlgorithmMaglev          LBAlgorithm = "Maglev"
	LBAlgorithmRoundRobin      LBAlgorithm = "RoundRobin"
	LBAlgorithmLeastConn       LBAlgorithm = "LeastConnections"
	LBAlgorithmIPHash          LBAlgorithm = "IPHash"
	LBAlgorithmRandom          LBAlgorithm = "Random"
)

// Protocol defines the L4 protocol.
type Protocol uint8

const (
	ProtocolTCP  Protocol = 6
	ProtocolUDP  Protocol = 17
	ProtocolSCTP Protocol = 132
	ProtocolICMP Protocol = 1
)

// L4Addr is an L4 address (IP + port + protocol).
type L4Addr struct {
	IP       netip.Addr
	Port     uint16
	Protocol Protocol
}

// Backend represents a service backend endpoint.
type Backend struct {
	ID      BackendID
	Address L4Addr
	NodeIP  netip.Addr
	Weight  uint32
	State   BackendState
}

// BackendState is the health state of a service backend.
type BackendState uint8

const (
	BackendStateActive   BackendState = 0
	BackendStateDraining BackendState = 1
	BackendStateDown     BackendState = 2
)

// DropReason is an eBPF packet drop reason code for observability.
type DropReason uint32

const (
	DropReasonNone          DropReason = 0
	DropReasonPolicyDenied  DropReason = 1
	DropReasonCTMapFull     DropReason = 2
	DropReasonNoBackend     DropReason = 3
	DropReasonInvalidPacket DropReason = 4
	DropReasonEncryptFail   DropReason = 5
)

// String returns a human-readable drop reason.
func (d DropReason) String() string {
	switch d {
	case DropReasonPolicyDenied:
		return "POLICY_DENIED"
	case DropReasonCTMapFull:
		return "CT_MAP_FULL"
	case DropReasonNoBackend:
		return "NO_BACKEND"
	case DropReasonInvalidPacket:
		return "INVALID_PACKET"
	case DropReasonEncryptFail:
		return "ENCRYPT_FAIL"
	default:
		return "NONE"
	}
}
