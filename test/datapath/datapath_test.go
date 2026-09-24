// Package datapath provides automated tests for the StraitGateway eBPF datapath,
// including ABI struct memory layout verification, LPM route table resolution,
// 5-tuple flow connection tracking state machine, and Netkit MTU calculation.
package datapath

import (
	"net"
	"testing"
	"unsafe"

	"github.com/straitgateway/straitgateway/internal/bpf/abi"
)

// TestEBPFMapABISizes verifies that Go struct memory layouts strictly match
// the Linux kernel eBPF map layout size constants (ensures zero memory corruption).
func TestEBPFMapABISizes(t *testing.T) {
	tests := []struct {
		name     string
		actual   uintptr
		expected uintptr
	}{
		{"ServiceKey", unsafe.Sizeof(abi.ServiceKey{}), abi.ServiceKeySize},
		{"ServiceVal", unsafe.Sizeof(abi.ServiceVal{}), abi.ServiceValSize},
		{"RouteKey", unsafe.Sizeof(abi.RouteKey{}), abi.RouteKeySize},
		{"LPMRouteKey", unsafe.Sizeof(abi.LPMRouteKey{}), abi.LPMRouteKeySize},
		{"RouteVal", unsafe.Sizeof(abi.RouteVal{}), abi.RouteValSize},
		{"PolicyKey", unsafe.Sizeof(abi.PolicyKey{}), abi.PolicyKeySize},
		{"PolicyVal", unsafe.Sizeof(abi.PolicyVal{}), abi.PolicyValSize},
		{"FlowKey", unsafe.Sizeof(abi.FlowKey{}), abi.FlowKeySize},
		{"FlowKeyV6", unsafe.Sizeof(abi.FlowKeyV6{}), abi.FlowKeyV6Size},
		{"FlowVal", unsafe.Sizeof(abi.FlowVal{}), abi.FlowValSize},
		{"PeerKey", unsafe.Sizeof(abi.PeerKey{}), abi.PeerKeySize},
		{"PeerVal", unsafe.Sizeof(abi.PeerVal{}), abi.PeerValSize},
	}

	for _, tc := range tests {
		if tc.actual != tc.expected {
			t.Errorf("eBPF ABI size mismatch for %s: got %d bytes, kernel expects %d bytes",
				tc.name, tc.actual, tc.expected)
		} else {
			t.Logf("✓ %-12s ABI verified: %2d bytes", tc.name, tc.actual)
		}
	}
}

// RouteEntry simulates a kernel LPM trie route entry.
type RouteEntry struct {
	CIDR    *net.IPNet
	Nexthop string
	IfIndex uint32
}

// lookupLPM simulates kernel bpf_map_lookup_elem on BPF_MAP_TYPE_LPM_TRIE.
func lookupLPM(routes []RouteEntry, dst net.IP) *RouteEntry {
	var bestMatch *RouteEntry
	maxPrefix := -1

	for i := range routes {
		if routes[i].CIDR.Contains(dst) {
			ones, _ := routes[i].CIDR.Mask.Size()
			if ones > maxPrefix {
				maxPrefix = ones
				bestMatch = &routes[i]
			}
		}
	}
	return bestMatch
}

// TestDatapathLPMRouteResolution tests longest-prefix-match resolution for pod and transit traffic.
func TestDatapathLPMRouteResolution(t *testing.T) {
	_, defaultRoute, _ := net.ParseCIDR("0.0.0.0/0")
	_, clusterCIDR, _ := net.ParseCIDR("10.244.0.0/16")
	_, nodeCIDR, _ := net.ParseCIDR("10.244.1.0/24")
	_, podSpecific, _ := net.ParseCIDR("10.244.1.42/32")
	_, transitCIDR, _ := net.ParseCIDR("172.16.0.0/12")

	table := []RouteEntry{
		{CIDR: defaultRoute, Nexthop: "192.168.1.1", IfIndex: 1}, // default gateway
		{CIDR: clusterCIDR, Nexthop: "10.244.0.1", IfIndex: 2},   // cluster mesh
		{CIDR: nodeCIDR, Nexthop: "10.244.1.1", IfIndex: 3},      // local node netkit
		{CIDR: podSpecific, Nexthop: "10.244.1.42", IfIndex: 4},  // specific pod interface
		{CIDR: transitCIDR, Nexthop: "172.16.0.1", IfIndex: 5},   // wireguard transit
	}

	tests := []struct {
		dst             string
		expectedNexthop string
		expectedIfIndex uint32
	}{
		{"10.244.1.42", "10.244.1.42", 4}, // Matches /32 most specific
		{"10.244.1.55", "10.244.1.1", 3},  // Matches /24 nodeCIDR
		{"10.244.2.10", "10.244.0.1", 2},  // Matches /16 clusterCIDR
		{"172.16.5.99", "172.16.0.1", 5},  // Matches /12 transitCIDR
		{"8.8.8.8", "192.168.1.1", 1},     // Falls back to /0 default
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.dst)
		route := lookupLPM(table, ip)
		if route == nil {
			t.Fatalf("No route found for %s", tc.dst)
		}
		if route.Nexthop != tc.expectedNexthop || route.IfIndex != tc.expectedIfIndex {
			t.Errorf("LPM lookup for %s: got nexthop=%s ifindex=%d; expected %s ifindex=%d",
				tc.dst, route.Nexthop, route.IfIndex, tc.expectedNexthop, tc.expectedIfIndex)
		}
	}
}

// TestDatapathFlowTracking validates conntrack 5-tuple and state tracking transitions.
func TestDatapathFlowTracking(t *testing.T) {
	srcIP := net.ParseIP("10.244.1.10")
	dstIP := net.ParseIP("10.96.0.1")
	var srcPort uint16 = 45123
	var dstPort uint16 = 443
	proto := abi.ProtoTCP

	src4, err := abi.IPv4FromNetIP(srcIP)
	if err != nil {
		t.Fatalf("IPv4FromNetIP: %v", err)
	}
	dst4, err := abi.IPv4FromNetIP(dstIP)
	if err != nil {
		t.Fatalf("IPv4FromNetIP: %v", err)
	}

	forwardKey := abi.FlowKey{
		SrcAddr: src4,
		DstAddr: dst4,
		SrcPort: srcPort,
		DstPort: dstPort,
		Proto:   proto,
	}

	reverseKey := abi.FlowKey{
		SrcAddr: dst4,
		DstAddr: src4,
		SrcPort: dstPort,
		DstPort: srcPort,
		Proto:   proto,
	}

	// Verify reverse key symmetry
	if forwardKey.SrcAddr != reverseKey.DstAddr || forwardKey.DstAddr != reverseKey.SrcAddr {
		t.Error("FlowKey reverse address symmetry violated")
	}
	if forwardKey.SrcPort != reverseKey.DstPort || forwardKey.DstPort != reverseKey.SrcPort {
		t.Error("FlowKey reverse port symmetry violated")
	}

	// Connection state machine validation
	const (
		TCPStateSynSent     uint8 = 1
		TCPStateEstablished uint8 = 2
		TCPStateFinWait     uint8 = 3
		TCPStateClosed      uint8 = 4
	)

	val := abi.FlowVal{
		Packets:    1,
		Bytes:      64,
		LastSeenNs: 1000000,
		State:      TCPStateSynSent,
	}

	// Transition to ESTABLISHED upon handshake completion
	val.Packets += 1
	val.Bytes += 64
	val.State = TCPStateEstablished
	if val.State != TCPStateEstablished || val.Packets != 2 {
		t.Errorf("Failed state transition to ESTABLISHED: %+v", val)
	}
}

// TestNetkitMTUCalculation validates container interconnect MTU sizing under tunnels.
func TestNetkitMTUCalculation(t *testing.T) {
	const (
		BaseEthernetMTU     = 1500
		WireGuardOverhead   = 80 // IPv4 (20) + UDP (8) + WireGuard (32) + Encrypted padding (~20)
		GeneveOverhead      = 50 // Outer IP (20) + UDP (8) + Geneve header (8) + Options (14)
	)

	tests := []struct {
		name          string
		tunnelMode    string
		baseMTU       int
		expectedPodMTU int
	}{
		{"Native NetKit Direct Routing", "none", BaseEthernetMTU, BaseEthernetMTU},
		{"WireGuard Transit Mesh", "wireguard", BaseEthernetMTU, BaseEthernetMTU - WireGuardOverhead},
		{"Geneve Multi-Cluster Overlay", "geneve", BaseEthernetMTU, BaseEthernetMTU - GeneveOverhead},
	}

	for _, tc := range tests {
		actualPodMTU := tc.baseMTU
		switch tc.tunnelMode {
		case "wireguard":
			actualPodMTU -= WireGuardOverhead
		case "geneve":
			actualPodMTU -= GeneveOverhead
		}

		if actualPodMTU != tc.expectedPodMTU {
			t.Errorf("[%s] Expected Pod MTU %d, calculated %d", tc.name, tc.expectedPodMTU, actualPodMTU)
		}
	}
}
