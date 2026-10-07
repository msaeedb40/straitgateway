package routing

import (
	"net"
	"testing"
)

func TestFIBEngineLPMAndPrecedence(t *testing.T) {
	fib := NewFIBEngine()

	// Create custom VRF 100
	fib.AddVRF(100, "tenant-a", "Tenant A isolated domain")

	_, dst1, _ := net.ParseCIDR("10.0.0.0/16")
	_, dst2, _ := net.ParseCIDR("10.0.1.0/24")

	// 1. Insert less specific route (/16) with StaticRoute precedence
	entry1 := &RouteEntry{
		Destination: *dst1,
		VRFID:       0,
		NextHopIP:   net.ParseIP("192.168.1.1"),
		Precedence:  PrecedenceStaticRoute,
		Metric:      10,
	}
	_, err := fib.InsertRoute(entry1)
	if err != nil {
		t.Fatalf("failed to insert route 1: %v", err)
	}

	// 2. Insert more specific route (/24)
	entry2 := &RouteEntry{
		Destination: *dst2,
		VRFID:       0,
		NextHopIP:   net.ParseIP("192.168.1.2"),
		Precedence:  PrecedenceBGPRoute,
		Metric:      20,
	}
	_, err = fib.InsertRoute(entry2)
	if err != nil {
		t.Fatalf("failed to insert route 2: %v", err)
	}

	// 3. Test LPM: 10.0.1.50 should match 10.0.1.0/24
	matched, err := fib.LongestPrefixMatch(0, net.ParseIP("10.0.1.50"))
	if err != nil {
		t.Fatalf("LPM match failed: %v", err)
	}
	if matched.Destination.String() != "10.0.1.0/24" {
		t.Errorf("expected match 10.0.1.0/24, got %s", matched.Destination.String())
	}

	// 4. Test LPM: 10.0.2.50 should match 10.0.0.0/16
	matched2, err := fib.LongestPrefixMatch(0, net.ParseIP("10.0.2.50"))
	if err != nil {
		t.Fatalf("LPM match failed: %v", err)
	}
	if matched2.Destination.String() != "10.0.0.0/16" {
		t.Errorf("expected match 10.0.0.0/16, got %s", matched2.Destination.String())
	}

	// 5. VRF Isolation: IP in VRF 100 should not resolve route in VRF 0
	_, err = fib.LongestPrefixMatch(100, net.ParseIP("10.0.1.50"))
	if err == nil {
		t.Error("expected error for unrouted IP in VRF 100, got nil")
	}

	// 6. Precedence test: inferior route with lower precedence should not overwrite
	inferiorEntry := &RouteEntry{
		Destination: *dst1,
		VRFID:       0,
		NextHopIP:   net.ParseIP("192.168.1.99"),
		Precedence:  PrecedenceDefaultRoute, // Lower priority
		Metric:      100,
	}
	inserted, _ := fib.InsertRoute(inferiorEntry)
	if inserted {
		t.Error("inferior route was accepted over higher precedence route")
	}
}
