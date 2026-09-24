// Package abi defines the explicit, versioned Go ↔ eBPF memory layouts
// and serialization contracts for StraitGateway eBPF maps and programs.
package abi

import "fmt"

// ABI size constants for flow maps
const (
	FlowKeySize   = 16
	FlowKeyV6Size = 40
	FlowValSize   = 40
)

// FlowKey matches `struct sg_flow_key` in bpf/include/sg_common.h (size: 16 bytes).
type FlowKey struct {
	SrcAddr IPv4Addr // Source IPv4 (network byte order)
	DstAddr IPv4Addr // Destination IPv4 (network byte order)
	SrcPort uint16   // Source port (network byte order)
	DstPort uint16   // Destination port (network byte order)
	Proto   uint8    // Protocol (IPPROTO_TCP, IPPROTO_UDP, etc.)
	Pad     [3]byte  // Alignment padding to 4 bytes
}

// FlowKeyV6 matches `struct sg_flow_v6_key` in bpf/include/sg_common.h (size: 40 bytes).
type FlowKeyV6 struct {
	SrcAddr IPv6Addr // Source IPv6 (network byte order)
	DstAddr IPv6Addr // Destination IPv6 (network byte order)
	SrcPort uint16   // Source port (network byte order)
	DstPort uint16   // Destination port (network byte order)
	Proto   uint8    // Protocol
	Pad     [3]byte  // Alignment padding
}

// FlowVal matches `struct sg_flow_val` in bpf/include/sg_common.h (size: 40 bytes).
type FlowVal struct {
	Packets    uint64 // Cumulative packets
	Bytes      uint64 // Cumulative bytes
	LastSeenNs uint64 // Nanoseconds timestamp (ktime_get_ns)
	SrcID      uint32 // Source identity
	DstID      uint32 // Destination identity
	State      uint8  // Connection / TCP state
	Pad        [7]byte// Struct alignment padding to 8 bytes
}

func (k FlowKey) String() string {
	return fmt.Sprintf("flow[%s:%d -> %s:%d proto=%d]", k.SrcAddr, k.SrcPort, k.DstAddr, k.DstPort, k.Proto)
}

func (k FlowKeyV6) String() string {
	return fmt.Sprintf("flow6[%s:%d -> %s:%d proto=%d]", k.SrcAddr, k.SrcPort, k.DstAddr, k.DstPort, k.Proto)
}

func (v FlowVal) String() string {
	return fmt.Sprintf("flow-val[pkts=%d bytes=%d last=%d src=%d dst=%d state=%d]",
		v.Packets, v.Bytes, v.LastSeenNs, v.SrcID, v.DstID, v.State)
}
