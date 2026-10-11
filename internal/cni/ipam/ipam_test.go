package ipam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIPAMAllocationLifecycle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ipam-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	statePath := filepath.Join(tmpDir, "ipam.json")

	cfg := Config{
		SubnetV4:  "10.244.1.0/24",
		SubnetV6:  "fd00:10:244:1::/64",
		StateFile: statePath,
	}

	ipam, err := NewIPAM(cfg)
	if err != nil {
		t.Fatalf("failed to create IPAM: %v", err)
	}

	// 1. Allocate for container-1
	v4_1, v6_1, err := ipam.Allocate("container-1")
	if err != nil {
		t.Fatalf("allocate container-1 failed: %v", err)
	}
	if v4_1.IP.String() != "10.244.1.2" {
		t.Errorf("expected first pod IP 10.244.1.2, got %s", v4_1.IP.String())
	}
	if v6_1 == nil {
		t.Error("expected non-nil IPv6 allocation")
	}

	// Test IsAllocated
	if !ipam.IsAllocated(v4_1.IP) {
		t.Errorf("expected %s to be reported as allocated", v4_1.IP)
	}

	// 2. Allocate for container-2
	v4_2, _, err := ipam.Allocate("container-2")
	if err != nil {
		t.Fatalf("allocate container-2 failed: %v", err)
	}
	if v4_2.IP.String() != "10.244.1.3" {
		t.Errorf("expected second pod IP 10.244.1.3, got %s", v4_2.IP.String())
	}

	// Test ListAllocations
	allocs := ipam.ListAllocations()
	if len(allocs) != 2 {
		t.Errorf("expected 2 allocations, got %d", len(allocs))
	}

	// 3. Idempotency test: allocate same container ID again returns same IP
	v4_repeat, _, err := ipam.Allocate("container-1")
	if err != nil {
		t.Fatalf("idempotent allocation failed: %v", err)
	}
	if v4_repeat.IP.String() != v4_1.IP.String() {
		t.Errorf("idempotent allocation gave different IP: got %s, expected %s", v4_repeat.IP.String(), v4_1.IP.String())
	}

	// 4. Test GarbageCollect: active containers only container-2
	activeMap := map[string]struct{}{
		"container-2": {},
	}
	reaped, err := ipam.GarbageCollect(activeMap)
	if err != nil {
		t.Fatalf("garbage collection error: %v", err)
	}
	if len(reaped) != 1 || reaped[0] != "container-1" {
		t.Errorf("expected container-1 reaped, got %v", reaped)
	}

	// Verify container-1 is no longer allocated
	if ipam.IsAllocated(v4_1.IP) {
		t.Errorf("expected %s to be freed by GC", v4_1.IP)
	}

	// 5. Reload from disk and verify persistence
	ipamReloaded, err := NewIPAM(cfg)
	if err != nil {
		t.Fatalf("failed to reload IPAM: %v", err)
	}

	// container-1 should now be re-allocatable with 10.244.1.2
	v4_realloc, _, err := ipamReloaded.Allocate("container-3")
	if err != nil {
		t.Fatalf("allocate container-3 failed: %v", err)
	}
	if v4_realloc.IP.String() != "10.244.1.2" {
		t.Errorf("expected reused IP 10.244.1.2, got %s", v4_realloc.IP.String())
	}
}
