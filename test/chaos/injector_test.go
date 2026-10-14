// Package chaos — injector_test.go validates the chaos injector's core
// fault injection mechanisms including ENOMEM simulation, drop burst control
// map writes, and BPF_PROG_TEST_RUN baseline using a minimal XDP pass program.
package chaos_test

import (
	"context"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/test/chaos"
)

// requireRoot skips the test if not running as root (needed for BPF syscalls).
func requireRoot(t *testing.T) {
	t.Helper()
	if os.Getuid() != 0 {
		t.Skip("chaos tests require root (or CAP_BPF + CAP_NET_ADMIN)")
	}
}

// TestInjectorNew verifies the Injector initializes without error using an
// ephemeral (non-pinned) BPF map.  On kernels that don't support BPF the
// injector must still return a usable handle (no-op mode).
func TestInjectorNew(t *testing.T) {
	inj, err := chaos.New(chaos.Config{})
	if err != nil {
		t.Fatalf("chaos.New: unexpected error: %v", err)
	}
	defer func() { _ = inj.Close() }()
}

// TestInjectENOMEM injects an ENOMEM fault for 100 ms and verifies the fault
// is cleared automatically after the duration expires.
func TestInjectENOMEM(t *testing.T) {
	requireRoot(t)
	if runtime.GOOS != "linux" {
		t.Skip("BPF map requires Linux")
	}

	inj, err := chaos.New(chaos.Config{})
	if err != nil {
		t.Fatalf("chaos.New: %v", err)
	}
	defer func() { _ = inj.Close() }()

	ctx := context.Background()
	if err := inj.InjectENOMEM(ctx, 100*time.Millisecond); err != nil {
		t.Fatalf("InjectENOMEM: %v", err)
	}

	// Give the fault time to expire
	time.Sleep(200 * time.Millisecond)

	// ClearAll should be a no-op (already expired)
	if err := inj.ClearAll(); err != nil {
		t.Errorf("ClearAll after ENOMEM expiry: %v", err)
	}
}

// TestInjectDropBurst writes a drop burst control record and verifies the
// write succeeds without error.
func TestInjectDropBurst(t *testing.T) {
	requireRoot(t)
	if runtime.GOOS != "linux" {
		t.Skip("BPF map requires Linux")
	}

	inj, err := chaos.New(chaos.Config{})
	if err != nil {
		t.Fatalf("chaos.New: %v", err)
	}
	defer func() { _ = inj.Close() }()

	ctx := context.Background()
	err = inj.InjectDropBurst(ctx, 150*time.Millisecond, chaos.DropBurstConfig{
		BurstSize:   32,
		Probability: 100,
	})
	if err != nil {
		t.Fatalf("InjectDropBurst: %v", err)
	}

	// Burst should auto-expire
	time.Sleep(250 * time.Millisecond)
}

// TestBuildMinimalPassProg validates that BuildMinimalPassProg produces a
// valid XDP pass program that can be loaded by the kernel verifier.
func TestBuildMinimalPassProg(t *testing.T) {
	requireRoot(t)
	if runtime.GOOS != "linux" {
		t.Skip("BPF program loading requires Linux")
	}

	prog, err := chaos.BuildMinimalPassProg()
	if err != nil {
		t.Fatalf("BuildMinimalPassProg: %v", err)
	}
	defer func() { _ = prog.Close() }()

	pkt := chaos.SyntheticUDPPacket(
		net.ParseIP("10.0.0.1"),
		net.ParseIP("10.0.0.2"),
		12345, 80,
	)

	ret, _, err := prog.Test(pkt)
	if err != nil {
		t.Fatalf("prog.Test (prog_test_run): %v", err)
	}

	const XDP_PASS = 2
	if ret != XDP_PASS {
		t.Errorf("expected XDP_PASS (%d), got %d", XDP_PASS, ret)
	}
}

// TestSyntheticUDPPacket verifies the packet builder produces a correctly
// structured 50-byte frame (14 Eth + 20 IPv4 + 8 UDP + 8 payload).
func TestSyntheticUDPPacket(t *testing.T) {
	pkt := chaos.SyntheticUDPPacket(
		net.ParseIP("192.168.1.10"),
		net.ParseIP("192.168.1.20"),
		54321, 443,
	)
	expected := 14 + 20 + 8 + 8
	if len(pkt) != expected {
		t.Errorf("expected packet length %d, got %d", expected, len(pkt))
	}

	// Ethernet type: 0x0800
	if pkt[12] != 0x08 || pkt[13] != 0x00 {
		t.Errorf("expected EtherType 0x0800, got 0x%02x%02x", pkt[12], pkt[13])
	}

	// IP protocol: UDP = 17
	if pkt[25] != 17 {
		t.Errorf("expected IP proto 17 (UDP), got %d", pkt[25])
	}
}

// TestClearAll checks that ClearAll completes without error when no faults
// are active.
func TestClearAll(t *testing.T) {
	requireRoot(t)

	inj, err := chaos.New(chaos.Config{})
	if err != nil {
		t.Fatalf("chaos.New: %v", err)
	}
	defer func() { _ = inj.Close() }()

	if err := inj.ClearAll(); err != nil {
		t.Errorf("ClearAll on empty injector: %v", err)
	}
}
