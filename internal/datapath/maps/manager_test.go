package maps

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/cilium/ebpf/rlimit"
	"github.com/straitgateway/straitgateway/pkg/types"
)

func TestMapManagerLifecycle(t *testing.T) {
	// Remove memlock limit if possible
	_ = rlimit.RemoveMemlock()

	cfg := DefaultConfig()
	mm, err := NewMapManager(cfg)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || isPermissionOrCapError(err) {
			t.Logf("Skipping BPF map creation: test running unprivileged without CAP_BPF: %v", err)
			return
		}
		t.Fatalf("failed to create MapManager: %v", err)
	}
	defer func() { _ = mm.Close() }()

	// Test Route Map CRUD (LPM Trie)
	lpmKey, err := types.NewLPMV4Key(net.ParseIP("10.244.1.0"), 24)
	if err != nil {
		t.Fatalf("failed to create LPM key: %v", err)
	}

	nextHopID := uint32(42)
	if err := mm.AddRoute(lpmKey, nextHopID); err != nil {
		t.Fatalf("failed to add route: %v", err)
	}

	gotNextHop, err := mm.LookupRoute(lpmKey)
	if err != nil {
		t.Fatalf("failed to lookup route: %v", err)
	}
	if gotNextHop != nextHopID {
		t.Errorf("expected next hop %d, got %d", nextHopID, gotNextHop)
	}

	// Test NextHop Map CRUD
	nhInfo := types.NextHopInfo{
		IfIndex: 2,
		IP:      0x0a000001,
		Flags:   1,
		VRFID:   0,
	}
	if err := mm.AddNextHop(nextHopID, nhInfo); err != nil {
		t.Fatalf("failed to add nexthop: %v", err)
	}

	// Test Policy Map CRUD
	polKey := types.PolicyKey{
		SrcID:   1001,
		DstID:   2001,
		DstPort: 8080,
		Proto:   6,
	}
	if err := mm.SetPolicy(polKey, types.PolicyAllow); err != nil {
		t.Fatalf("failed to set policy: %v", err)
	}

	// Test Identity Map CRUD
	testIP := uint32(0x0a000105) // 10.0.1.5
	testID := uint32(500)
	if err := mm.SetIdentity(testIP, testID); err != nil {
		t.Fatalf("failed to set identity: %v", err)
	}
	gotID, err := mm.LookupIdentity(testIP)
	if err != nil || gotID != testID {
		t.Fatalf("lookup identity: expected %d, got %d, err: %v", testID, gotID, err)
	}
	if err := mm.DeleteIdentity(testIP); err != nil {
		t.Fatalf("delete identity: %v", err)
	}

	// Test VRF Map CRUD
	vrfCfg := types.VRFConfig{ID: 100, TableID: 1001, Flags: 1}
	if err := mm.SetVRF(100, vrfCfg); err != nil {
		t.Fatalf("failed to set VRF: %v", err)
	}
	gotVRF, err := mm.LookupVRF(100)
	if err != nil || gotVRF.TableID != 1001 {
		t.Fatalf("lookup VRF: expected table 1001, got %d, err: %v", gotVRF.TableID, err)
	}
	if err := mm.DeleteVRF(100); err != nil {
		t.Fatalf("delete VRF: %v", err)
	}

	// Test Tunnel Map CRUD
	tunInfo := types.TunnelInfo{TunnelID: 1, IfIndex: 10, EndpointIP: 0x0a0a0a01, EndpointPort: 51820}
	if err := mm.SetTunnel(1, tunInfo); err != nil {
		t.Fatalf("failed to set tunnel: %v", err)
	}
	gotTun, err := mm.LookupTunnel(1)
	if err != nil || gotTun.EndpointPort != 51820 {
		t.Fatalf("lookup tunnel: expected port 51820, got %d, err: %v", gotTun.EndpointPort, err)
	}
	if err := mm.DeleteTunnel(1); err != nil {
		t.Fatalf("delete tunnel: %v", err)
	}

	// Delete and verify
	if err := mm.DeleteRoute(lpmKey); err != nil {
		t.Fatalf("failed to delete route: %v", err)
	}

	_, err = mm.LookupRoute(lpmKey)
	if err == nil {
		t.Fatal("expected error after route deletion, got nil")
	}
}

func isPermissionOrCapError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errors.Is(err, os.ErrPermission) ||
		containsSubstring(errStr, "operation not permitted") ||
		containsSubstring(errStr, "permission denied")
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || searchSub(s, sub))
}

func searchSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
