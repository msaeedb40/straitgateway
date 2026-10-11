package routing

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/internal/routing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LocalASN != 65001 {
		t.Errorf("expected default LocalASN 65001, got %d", cfg.LocalASN)
	}
	if cfg.WireGuardDev != "wg0" {
		t.Errorf("expected default WireGuardDev wg0, got %s", cfg.WireGuardDev)
	}
}

func TestRoutingDataplaneLifecycle(t *testing.T) {
	cfg := DefaultConfig()
	mgr, err := NewManager(cfg, nil)
	if err != nil {
		t.Fatalf("failed to create routing manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	if mgr.FIB() == nil {
		t.Errorf("expected non-nil FIB engine")
	}
	if mgr.BGPSpeaker() == nil {
		t.Errorf("expected non-nil BGP speaker")
	}
	if mgr.BFDMonitor() == nil {
		t.Errorf("expected non-nil BFD monitor")
	}

	// 1. VRF testing
	if err := mgr.AddVRF(100, "tenant-a", "Tenant A Isolated VRF"); err != nil {
		t.Errorf("failed to add VRF: %v", err)
	}

	// 2. Route installation testing
	_, cidr, _ := net.ParseCIDR("10.200.0.0/16")
	nextHop := net.ParseIP("192.168.10.1")

	err = mgr.InstallRoute(*cidr, nextHop, 100, "eth0", routing.PrecedenceBGPRoute)
	if err != nil {
		t.Errorf("failed to install route: %v", err)
	}

	route, err := mgr.FIB().LongestPrefixMatch(100, net.ParseIP("10.200.5.10"))
	if err != nil || route == nil || !route.NextHopIP.Equal(nextHop) {
		t.Errorf("expected route lookup to resolve to %s: %v", nextHop, err)
	}

	// 3. BFD Failover testing
	session := mgr.SetupBFDPeering(nextHop, 50, 50)
	if session == nil {
		t.Fatalf("expected non-nil BFD session")
	}

	// Trigger BFD failure -> should withdraw route pointing to nextHop
	if err := mgr.BFDMonitor().TriggerFailure(nextHop); err != nil {
		t.Errorf("failed to trigger BFD failure: %v", err)
	}

	// Route should be withdrawn
	routeAfterFail, _ := mgr.FIB().LongestPrefixMatch(100, net.ParseIP("10.200.5.10"))
	if routeAfterFail != nil {
		t.Errorf("expected route to be withdrawn on BFD peer failure")
	}
}
