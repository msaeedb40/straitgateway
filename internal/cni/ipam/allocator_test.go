package ipam_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
)

func TestAllocate(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"10.244.1.0/24"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	// First allocation skips .0 (network) and .1 (gateway).
	cidr, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if !strings.HasPrefix(cidr, "10.244.1.") {
		t.Errorf("Allocate() = %q, expected prefix 10.244.1.", cidr)
	}
	if cidr == "10.244.1.0/24" {
		t.Error("allocated network address — should be skipped")
	}
	if cidr == "10.244.1.1/24" {
		t.Error("allocated gateway address — should be skipped")
	}
}

func TestAllocateDualStack(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"10.244.2.0/24", "fd00:10:244:2::/64"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	if !alloc.IsDualStack() {
		t.Errorf("expected IsDualStack() to be true")
	}

	allocation, err := alloc.AllocateDualStack()
	if err != nil {
		t.Fatalf("AllocateDualStack: %v", err)
	}

	if allocation.IPv4 == "" || allocation.IPv6 == "" {
		t.Errorf("Dual stack allocation incomplete: v4=%s, v6=%s", allocation.IPv4, allocation.IPv6)
	}
	if !strings.HasPrefix(allocation.IPv4, "10.244.2.") {
		t.Errorf("expected IPv4 in 10.244.2.0/24, got %s", allocation.IPv4)
	}
	if !strings.HasPrefix(allocation.IPv6, "fd00:10:244:2::") {
		t.Errorf("expected IPv6 in fd00:10:244:2::/64, got %s", allocation.IPv6)
	}
	if alloc.Gateway4().String() != "10.244.2.1" {
		t.Errorf("Gateway4: got %s, want 10.244.2.1", alloc.Gateway4().String())
	}
	if alloc.Gateway6().String() != "fd00:10:244:2::1" {
		t.Errorf("Gateway6: got %s, want fd00:10:244:2::1", alloc.Gateway6().String())
	}

	// Release IPv4 and IPv6
	if err := alloc.Release(allocation.IPv4); err != nil {
		t.Errorf("Release IPv4 failed: %v", err)
	}
	if err := alloc.Release(allocation.IPv6); err != nil {
		t.Errorf("Release IPv6 failed: %v", err)
	}
}

func TestAllocateAndRelease(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"192.168.42.0/28"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	allocated := make([]string, 0)
	// /28 = 16 addresses, 1 network + 1 gateway + 1 broadcast = 13 usable
	for i := 0; i < 13; i++ {
		cidr, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("Allocate #%d: %v", i, err)
		}
		allocated = append(allocated, cidr)
	}

	// Pool should now be exhausted
	_, err = alloc.Allocate()
	if err == nil {
		t.Error("expected error when pool exhausted, got nil")
	}

	// Release one and re-allocate
	if err := alloc.Release(allocated[0]); err != nil {
		t.Fatalf("Release: %v", err)
	}
	cidr, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocate after Release: %v", err)
	}
	if cidr != allocated[0] {
		t.Errorf("expected re-allocation of %q, got %q", allocated[0], cidr)
	}
}

func TestReservationAndRestoration(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"10.244.3.0/24"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	// Manually reserve 10.244.3.2/24 (which would be first allocated)
	if err := alloc.Reserve("10.244.3.2/24"); err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}

	// Next allocation should be 10.244.3.3/24
	cidr, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}
	if cidr != "10.244.3.3/24" {
		t.Errorf("got %s, want 10.244.3.3/24", cidr)
	}
}

func TestConcurrentAllocations(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"10.244.0.0/20"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	var wg sync.WaitGroup
	numWorkers := 20
	allocsPerWorker := 50
	allocatedMap := sync.Map{}

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < allocsPerWorker; i++ {
				ip, err := alloc.Allocate()
				if err != nil {
					t.Errorf("concurrent allocate failed: %v", err)
					return
				}
				if _, loaded := allocatedMap.LoadOrStore(ip, true); loaded {
					t.Errorf("duplicate IP allocated: %s", ip)
				}
			}
		}()
	}
	wg.Wait()
}

func TestGateway(t *testing.T) {
	alloc, err := ipam.NewAllocator([]string{"172.16.10.0/24"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}
	gw := alloc.Gateway()
	if gw.String() != "172.16.10.1" {
		t.Errorf("Gateway() = %q, want 172.16.10.1", gw.String())
	}
}

func TestNoCIDRError(t *testing.T) {
	_, err := ipam.NewAllocator(nil)
	if err == nil {
		t.Error("expected error for nil CIDRs")
	}
}

func TestExhaustionCounters(t *testing.T) {
	// A /30 IPv4 CIDR has 4 IPs total: .0 (network), .1 (gateway), .2 (usable host), .3 (broadcast)
	alloc, err := ipam.NewAllocator([]string{"192.168.1.0/30"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	ip, err := alloc.AllocateV4()
	if err != nil {
		t.Fatalf("AllocateV4() unexpected error: %v", err)
	}
	if ip != "192.168.1.2/30" {
		t.Fatalf("expected 192.168.1.2/30, got %s", ip)
	}

	// Next allocation must exhaust the pool and increment ExhaustionEventsV4
	_, err = alloc.AllocateV4()
	if err == nil {
		t.Fatalf("expected pool exhaustion error, got nil")
	}
	if alloc.ExhaustionEventsV4 != 1 {
		t.Errorf("expected ExhaustionEventsV4 == 1, got %d", alloc.ExhaustionEventsV4)
	}
}

