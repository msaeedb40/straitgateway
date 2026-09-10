// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package maps provides typed Go wrappers around pinned BPF maps.
// These wrappers enforce type safety at the Go level for keys/values
// that correspond to the C struct definitions in bpf/maps/maps.h.
//
// All map operations go through the dataplane compiler — direct map
// writes from controllers are architecturally prohibited.
package maps

import (
	"net/netip"
	"unsafe"
)

// --- Service Map (service_key → service_value) ---

// ServiceKey matches the C struct service_key in maps.h.
type ServiceKey struct {
	VIP4     [4]byte // network byte order
	Port     uint16
	Protocol uint8
	Pad      uint8
}

// ServiceValue matches the C struct service_value in maps.h.
type ServiceValue struct {
	BackendCount uint32
	Algorithm    uint8  // 0=maglev, 1=roundrobin, 2=random
	Flags        uint8  // bit0=DSR, bit1=sessionAffinity, bit2=nodePort
	Pad          [2]uint8
}

// --- Backend Map (backend_key → backend_value) ---

// BackendKey matches the C struct backend_key in maps.h.
type BackendKey struct {
	ID uint32
}

// BackendValue matches the C struct backend_value in maps.h.
type BackendValue struct {
	IP4      [4]byte
	Port     uint16
	Protocol uint8
	State    uint8  // 0=active, 1=terminating, 2=quarantined
	Weight   uint32
}

// --- Policy Map (policy_key → policy_value) ---

// PolicyKey matches the C struct policy_key in maps.h.
type PolicyKey struct {
	SrcIdentity uint32
	DstIdentity uint32
	DstPort     uint16
	Protocol    uint8
	Pad         uint8
}

// PolicyValue matches the C struct policy_value in maps.h.
type PolicyValue struct {
	Action   uint8  // 0=allow, 1=deny, 2=reject
	Priority uint8
	Pad      [2]uint8
}

// --- Identity Map (identity_key → identity_value) ---

// IdentityKey matches the C struct identity_key in maps.h.
type IdentityKey struct {
	Identity uint32
}

// IdentityValue matches the C struct identity_value in maps.h.
type IdentityValue struct {
	LabelsHash uint64
	SegmentID  uint32
	Pad        uint32
}

// --- Conntrack Map (ct_key → ct_value) ---

// ConntrackKey matches the C struct ct_key in maps.h.
type ConntrackKey struct {
	SrcIP4   [4]byte
	DstIP4   [4]byte
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
	Pad      [3]uint8
}

// ConntrackValue matches the C struct ct_value in maps.h.
type ConntrackValue struct {
	Packets   uint64
	Bytes     uint64
	Lifetime  uint32 // seconds
	Flags     uint8
	State     uint8  // 0=new, 1=established, 2=related, 3=closing
	Pad       [2]uint8
}

// --- Node Map (node_key → node_value) ---

// NodeKey matches the C struct node_key in maps.h.
type NodeKey struct {
	IP4 [4]byte
}

// NodeValue matches the C struct node_value in maps.h.
type NodeValue struct {
	WGPubKey  [32]byte
	TunnelIP4 [4]byte
	SegmentID uint32
}

// --- Helper Functions ---

// AddrToIPv4 converts a netip.Addr to a 4-byte array in network byte order.
func AddrToIPv4(addr netip.Addr) [4]byte {
	if !addr.Is4() {
		return [4]byte{}
	}
	return addr.As4()
}

// SizeOf returns the byte size of a struct value (unsafe).
func SizeOf[T any]() int {
	var zero T
	return int(unsafe.Sizeof(zero))
}
