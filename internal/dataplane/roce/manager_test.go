// Package roce — manager_test.go provides unit tests for the RoCEv2 manager,
// DCQCN rate controller, and NCCL topology engine. These tests are designed
// to run without root privileges or actual GPU/RDMA hardware.
package roce_test

import (
	"context"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/dataplane/roce"
)

// ---------------------------------------------------------------------------
// QP state machine tests
// ---------------------------------------------------------------------------

func TestQPStateMachineValidTransitions(t *testing.T) {
	// RESET → INIT → RTR → RTS → SQD → ERR → RESET
	transitions := []struct {
		from   roce.QPState
		to     roce.QPState
		wantOk bool
	}{
		{roce.QPStateReset, roce.QPStateInit, true},
		{roce.QPStateInit, roce.QPStateRTR, true},
		{roce.QPStateRTR, roce.QPStateRTS, true},
		{roce.QPStateRTS, roce.QPStateSQD, true},
		{roce.QPStateSQD, roce.QPStateErr, true},
		{roce.QPStateErr, roce.QPStateReset, true},
		// Invalid transitions
		{roce.QPStateReset, roce.QPStateRTS, false},
		{roce.QPStateRTR, roce.QPStateInit, false},
		{roce.QPStateRTS, roce.QPStateRTR, false},
	}

	for _, tc := range transitions {
		mgr := newTestManager(t)
		// Register QP in the 'from' state
		qpn := uint32(42)
		entry := roce.QPEntry{
			State:       uint8(tc.from),
			PFCPriority: 3,
			MTU:         4096,
			RemoteIP:    ipToUint32(net.ParseIP("10.0.0.1")),
		}
		if err := mgr.UpsertQP(qpn, entry); err != nil {
			t.Fatalf("UpsertQP: %v", err)
		}

		err := mgr.TransitionQP(qpn, tc.to)
		if tc.wantOk && err != nil {
			t.Errorf("transition %s → %s: expected OK, got %v", tc.from, tc.to, err)
		}
		if !tc.wantOk && err == nil {
			t.Errorf("transition %s → %s: expected error, got nil", tc.from, tc.to)
		}
	}
}

func TestQPStateString(t *testing.T) {
	cases := []struct {
		state roce.QPState
		want  string
	}{
		{roce.QPStateReset, "RESET"},
		{roce.QPStateInit, "INIT"},
		{roce.QPStateRTR, "RTR"},
		{roce.QPStateRTS, "RTS"},
		{roce.QPStateSQD, "SQD"},
		{roce.QPStateSQE, "SQE"},
		{roce.QPStateErr, "ERR"},
		{roce.QPStateDrain, "DRAIN"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("QPState(%d).String() = %q, want %q", uint8(tc.state), got, tc.want)
		}
	}
}

func TestUpsertAndDeleteQP(t *testing.T) {
	mgr := newTestManager(t)
	qpn := uint32(100)

	entry := roce.QPEntry{
		State:       uint8(roce.QPStateRTS),
		PFCPriority: 3,
		MTU:         4096,
		RemoteIP:    ipToUint32(net.ParseIP("192.168.1.10")),
		RQPSN:       1000,
		SQPSN:       2000,
	}

	if err := mgr.UpsertQP(qpn, entry); err != nil {
		t.Fatalf("UpsertQP: %v", err)
	}

	if err := mgr.DeleteQP(qpn); err != nil {
		t.Fatalf("DeleteQP: %v", err)
	}

	// Transitioning a deleted QP should fail
	if err := mgr.TransitionQP(qpn, roce.QPStateErr); err == nil {
		t.Error("expected error transitioning deleted QP, got nil")
	}
}

// ---------------------------------------------------------------------------
// DCQCN rate controller tests
// ---------------------------------------------------------------------------

func TestDCQCNRateReductionOnCNP(t *testing.T) {
	cfg := roce.DefaultDCQCNConfig()
	cfg.MaxRateMbps = 10_000 // 10 Gbps for test
	cfg.MinRateMbps = 100
	cfg.Alpha = 1.0

	rc := roce.NewRateController(cfg)

	qpn := uint32(77)
	rc.RegisterQP(qpn)

	initialStats, _ := rc.QPStats(qpn)
	if initialStats.CurrentRateMbps != cfg.MaxRateMbps {
		t.Fatalf("initial rate = %.0f, want %.0f", initialStats.CurrentRateMbps, cfg.MaxRateMbps)
	}

	// Simulate CNP — with Alpha=1.0, rate should halve
	rc.OnCNPReceived(qpn)

	// Wait past the decrease interval
	time.Sleep(cfg.RateDecreaseInterval + 1*time.Millisecond)

	stats, _ := rc.QPStats(qpn)
	// Rate should have decreased: Rc = Rc × (1 - 1.0/2) = Rc/2
	expectedRate := cfg.MaxRateMbps * (1.0 - 1.0/2.0)
	if stats.CurrentRateMbps != expectedRate {
		t.Errorf("rate after CNP = %.2f, want %.2f", stats.CurrentRateMbps, expectedRate)
	}
	if stats.TotalCNPs != 1 {
		t.Errorf("TotalCNPs = %d, want 1", stats.TotalCNPs)
	}
}

func TestDCQCNRateIncreaseOnTimerExpiry(t *testing.T) {
	cfg := roce.DefaultDCQCNConfig()
	cfg.MaxRateMbps = 10_000
	cfg.MinRateMbps = 100
	cfg.RateAI = 500 // Large AI for easy testing
	cfg.TimerPeriod = 10 * time.Millisecond

	rc := roce.NewRateController(cfg)
	qpn := uint32(88)
	rc.RegisterQP(qpn)

	// First reduce the rate
	rc.OnCNPReceived(qpn)
	time.Sleep(cfg.RateDecreaseInterval + 5*time.Millisecond)

	statsBefore, _ := rc.QPStats(qpn)
	reducedRate := statsBefore.CurrentRateMbps

	// Simulate timer expiry with no new CNP (wait past TimerPeriod)
	time.Sleep(cfg.TimerPeriod + 5*time.Millisecond)
	rc.OnTimerExpiry(qpn, 100_000)

	statsAfter, _ := rc.QPStats(qpn)
	if statsAfter.CurrentRateMbps <= reducedRate {
		t.Errorf("rate did not increase after timer expiry: before=%.2f, after=%.2f",
			reducedRate, statsAfter.CurrentRateMbps)
	}
}

func TestDCQCNHAIMode(t *testing.T) {
	cfg := roce.DefaultDCQCNConfig()
	cfg.MaxRateMbps = 10_000
	cfg.RateAI = 100
	cfg.RateHAI = 1000 // Much larger than AI
	cfg.HAITimerCount = 3
	cfg.TimerPeriod = 5 * time.Millisecond

	rc := roce.NewRateController(cfg)
	qpn := uint32(99)
	rc.RegisterQP(qpn)

	// Reduce rate first
	rc.OnCNPReceived(qpn)
	time.Sleep(cfg.RateDecreaseInterval + 2*time.Millisecond)

	statsBefore, _ := rc.QPStats(qpn)
	reducedRate := statsBefore.CurrentRateMbps

	// Fire HAITimerCount+1 timer expiries without CNP
	for i := 0; i <= cfg.HAITimerCount; i++ {
		time.Sleep(cfg.TimerPeriod + 2*time.Millisecond)
		rc.OnTimerExpiry(qpn, 0)
	}

	statsAfter, _ := rc.QPStats(qpn)
	// After HAI kicks in, rate should have increased more than with pure AI
	expectedMinIncrease := cfg.RateHAI // At least one HAI step
	if statsAfter.CurrentRateMbps < reducedRate+expectedMinIncrease {
		t.Errorf("HAI: rate increase insufficient: before=%.0f, after=%.0f (want +%.0f minimum)",
			reducedRate, statsAfter.CurrentRateMbps, expectedMinIncrease)
	}
}

func TestDCQCNRateFloor(t *testing.T) {
	cfg := roce.DefaultDCQCNConfig()
	cfg.MaxRateMbps = 1000
	cfg.MinRateMbps = 100
	cfg.Alpha = 1.0
	cfg.RateDecreaseInterval = 0

	rc := roce.NewRateController(cfg)
	qpn := uint32(11)
	rc.RegisterQP(qpn)

	// Apply many CNPs to try to drive rate to zero
	for i := 0; i < 30; i++ {
		rc.OnCNPReceived(qpn)
		time.Sleep(1 * time.Millisecond)
	}

	stats, _ := rc.QPStats(qpn)
	if stats.CurrentRateMbps < cfg.MinRateMbps {
		t.Errorf("rate %.2f went below MinRateMbps %.2f", stats.CurrentRateMbps, cfg.MinRateMbps)
	}
}

func TestRateControllerDeregister(t *testing.T) {
	cfg := roce.DefaultDCQCNConfig()
	rc := roce.NewRateController(cfg)

	qpn := uint32(55)
	rc.RegisterQP(qpn)
	rc.DeregisterQP(qpn)

	_, err := rc.QPStats(qpn)
	if err == nil {
		t.Error("expected error for deregistered QPN, got nil")
	}
}

// ---------------------------------------------------------------------------
// NCCL topology tests
// ---------------------------------------------------------------------------

func TestTopologyGraphBuild(t *testing.T) {
	g := roce.NewTopologyGraph()
	if err := g.Build(); err != nil {
		t.Fatalf("TopologyGraph.Build: %v", err)
	}
	// Build should succeed even without nvidia-smi (falls back to synthetic endpoints)
}

func TestTopologyGraphBPFEntries(t *testing.T) {
	g := roce.NewTopologyGraph()
	_ = g.Build()

	entries := g.BPFEntries()
	// With synthetic endpoints, we should have at least some entries
	// (may be 0 if both endpoints have same GPUID — just verify no panic)
	t.Logf("BPF topology entries: %d", len(entries))
}

func TestEncodeGPUVirtualIP(t *testing.T) {
	// Verify the virtual IP encoding: 169.254.0.gpuID
	entries, err := roce.DiscoverGPUTopology()
	if err != nil {
		t.Fatalf("DiscoverGPUTopology: %v", err)
	}
	// Each returned IP should be in 169.254.0.0/16
	_, subnet, _ := net.ParseCIDR("169.254.0.0/16")
	for remoteIP := range entries {
		ip := uint32ToIP(remoteIP)
		if !subnet.Contains(ip) {
			t.Errorf("GPU virtual IP %v is not in 169.254.0.0/16", ip)
		}
	}
}

// ---------------------------------------------------------------------------
// Manager integration test (no BPF)
// ---------------------------------------------------------------------------

func TestManagerLifecycle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("RoCE Manager requires Linux")
	}

	cfg := roce.DefaultConfig("lo")
	cfg.XDPObjPath = "" // skip XDP loading — ELF not available in unit tests
	cfg.TCObjPath = ""  // skip TC loading
	cfg.BPFFSBase = t.TempDir() + "/roce_bpffs"

	mgr, err := roce.New(cfg)
	if err != nil {
		if os.Getuid() != 0 {
			t.Skipf("RoCE Manager requires root for BPF map creation: %v", err)
		}
		t.Fatalf("roce.New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Manager.Start: %v", err)
	}

	if err := mgr.Stop(); err != nil {
		t.Fatalf("Manager.Stop: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestManager(t *testing.T) *roce.Manager {
	t.Helper()
	cfg := roce.DefaultConfig("lo")
	cfg.XDPObjPath = ""
	cfg.TCObjPath = ""
	cfg.BPFFSBase = t.TempDir() + "/roce_bpffs"

	mgr, err := roce.New(cfg)
	if err != nil {
		if os.Getuid() != 0 {
			t.Skipf("BPF map creation requires root: %v", err)
		}
		t.Fatalf("roce.New: %v", err)
	}
	t.Cleanup(func() { mgr.Stop() })
	return mgr
}

func ipToUint32(ip net.IP) uint32 {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
}

func uint32ToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
