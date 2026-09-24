package abi

import (
	"fmt"
	"net"
)

// ABI size constants for eBPF map creation (Finding #12)
const (
	RouteKeySize    = 8
	LPMRouteKeySize = 8
	RouteValSize    = 12
)

// RouteKey matches `struct sg_route_key` in bpf/maps/routes.bpf.c (size: 8 bytes).
type RouteKey struct {
	Prefix    IPv4Addr // IPv4 network address in network order
	PrefixLen uint8    // Prefix length (0–32)
	Pad       [3]byte  // Alignment padding
}

// LPMRouteKey provides the standard BPF_MAP_TYPE_LPM_TRIE layout required by the Linux kernel,
// where prefixlen must be the first 32-bit field.
type LPMRouteKey struct {
	PrefixLen uint32   // Prefix length in bits (0-32)
	Prefix    IPv4Addr // IPv4 network address
}

// RouteVal matches `struct sg_route_val` in bpf/maps/routes.bpf.c (size: 12 bytes).
type RouteVal struct {
	Nexthop IPv4Addr // Next-hop IPv4 address
	IfIndex uint32   // Output interface index
	Flags   uint8    // RouteFlag* (RouteFlagBlackhole, RouteFlagLocal)
	Pad     [3]byte  // Alignment padding
}

// NewRouteKey creates a RouteKey from a CIDR.
func NewRouteKey(cidr *net.IPNet) (RouteKey, error) {
	ip4, err := IPv4FromNetIP(cidr.IP)
	if err != nil {
		return RouteKey{}, err
	}
	ones, _ := cidr.Mask.Size()
	return RouteKey{
		Prefix:    ip4,
		PrefixLen: uint8(ones),
		Pad:       [3]byte{},
	}, nil
}

// NewLPMRouteKey creates a standard kernel LPM trie key from a CIDR.
func NewLPMRouteKey(cidr *net.IPNet) (LPMRouteKey, error) {
	ip4, err := IPv4FromNetIP(cidr.IP)
	if err != nil {
		return LPMRouteKey{}, err
	}
	ones, _ := cidr.Mask.Size()
	return LPMRouteKey{
		PrefixLen: uint32(ones),
		Prefix:    ip4,
	}, nil
}

// RouteKeyV6 represents an IPv6 route key (size: 20 bytes).
type RouteKeyV6 struct {
	PrefixLen uint32   // Prefix length (0-128)
	Prefix    IPv6Addr // IPv6 network address
}

// RouteValV6 represents an IPv6 route value (size: 24 bytes).
type RouteValV6 struct {
	Nexthop IPv6Addr // Next-hop IPv6 address
	IfIndex uint32   // Output interface index
	Flags   uint8    // Flags
	Pad     [3]byte  // Alignment
}

func (k RouteKey) String() string {
	return fmt.Sprintf("%s/%d", k.Prefix, k.PrefixLen)
}
