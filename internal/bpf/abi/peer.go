// Package abi defines the explicit, versioned Go ↔ eBPF memory layouts
// and serialization contracts for StraitGateway eBPF maps and programs.
package abi

import "fmt"

// ABI size constants for peer maps
const (
	PeerKeySize   = 4
	PeerValSize   = 8
	PeerValV6Size = 20

	PeerTypeTransit uint8 = 1
	PeerTypeMesh    uint8 = 2
)

// PeerKey matches `struct sg_peer_key` in bpf/include/sg_common.h (size: 4 bytes).
type PeerKey struct {
	PeerID uint32 // Peer identity / node identifier
}

// PeerVal matches `struct sg_peer_val` in bpf/include/sg_common.h (size: 8 bytes).
type PeerVal struct {
	Addr  IPv4Addr // Peer IPv4 endpoint address (network byte order)
	Port  uint16   // Peer port (network byte order)
	Type  uint8    // PeerTypeTransit or PeerTypeMesh
	Flags uint8    // Peer state flags
}

// PeerValV6 matches `struct sg_peer_v6_val` in bpf/include/sg_common.h (size: 20 bytes).
type PeerValV6 struct {
	Addr  IPv6Addr // Peer IPv6 endpoint address (network byte order)
	Port  uint16   // Peer port (network byte order)
	Type  uint8    // PeerTypeTransit or PeerTypeMesh
	Flags uint8    // Peer state flags
}

func (k PeerKey) String() string {
	return fmt.Sprintf("peer[%d]", k.PeerID)
}

func (v PeerVal) String() string {
	return fmt.Sprintf("peer-val[%s:%d type=%d]", v.Addr, v.Port, v.Type)
}

func (v PeerValV6) String() string {
	return fmt.Sprintf("peer-val6[%s:%d type=%d]", v.Addr, v.Port, v.Type)
}
