package types

import (
	"encoding/binary"
	"fmt"
	"net"
)

// LPMV4Key represents an IPv4 Longest Prefix Match key for eBPF route lookup
type LPMV4Key struct {
	PrefixLen uint32
	Addr      uint32
}

// NewLPMV4Key creates an LPMV4Key from an IP and prefix length
func NewLPMV4Key(ip net.IP, prefixLen int) (LPMV4Key, error) {
	ip4 := ip.To4()
	if ip4 == nil {
		return LPMV4Key{}, fmt.Errorf("invalid IPv4 address: %v", ip)
	}
	return LPMV4Key{
		PrefixLen: uint32(prefixLen),
		Addr:      binary.BigEndian.Uint32(ip4),
	}, nil
}

// NextHopInfo holds the forwarding target metadata in eBPF
type NextHopInfo struct {
	IfIndex uint32
	IP      uint32
	MAC     [6]byte
	Flags   uint16
	VRFID   uint32
}

// ServiceKey represents a VIP service mapping key
type ServiceKey struct {
	IP    uint32
	Port  uint16
	Proto uint8
	Pad   uint8
}

// ServiceEndpoint holds a backend pod target
type ServiceEndpoint struct {
	IP     uint32
	Port   uint16
	Weight uint8
	State  uint8
}

// FlowKey represents a 5-tuple conntrack flow
type FlowKey struct {
	SrcIP   uint32
	DstIP   uint32
	SrcPort uint16
	DstPort uint16
	Proto   uint8
	Pad     [3]byte
}

// FlowEntry represents connection tracking state
type FlowEntry struct {
	Bytes       uint64
	Packets     uint64
	LastSeen    uint64
	Flags       uint32
	RevSrcIP    uint32
	RevDstIP    uint32
	RevSrcPort  uint16
	RevDstPort  uint16
}

// PolicyKey matches source and destination workload identities
type PolicyKey struct {
	SrcID   uint32
	DstID   uint32
	DstPort uint16
	Proto   uint8
	Pad     uint8
}

// PolicyAction defines allow, drop, or audit
type PolicyAction uint32

const (
	PolicyDrop  PolicyAction = 0
	PolicyAllow PolicyAction = 1
	PolicyAudit PolicyAction = 2
)
