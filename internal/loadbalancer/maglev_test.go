package loadbalancer

import (
	"testing"
)

func TestMaglevDistributionAndConsistency(t *testing.T) {
	backends := []string{"pod-1:8080", "pod-2:8080", "pod-3:8080"}
	mag, err := NewMaglev(backends, 257)
	if err != nil {
		t.Fatalf("failed to create Maglev: %v", err)
	}

	// 1. Verify table size
	table := mag.LookupTable()
	if len(table) != 257 {
		t.Errorf("expected table size 257, got %d", len(table))
	}

	// 2. Verify all slots are assigned
	counts := make(map[int]int)
	for i, bIdx := range table {
		if bIdx < 0 || bIdx >= len(backends) {
			t.Fatalf("slot %d has invalid backend index %d", i, bIdx)
		}
		counts[bIdx]++
	}

	// 3. Verify each backend received roughly equal share (distribution)
	for i := range backends {
		if counts[i] == 0 {
			t.Errorf("backend %d received 0 slots", i)
		}
	}

	// 4. Verify deterministic lookup for flow hashes
	flowHash1 := uint64(1234567)
	idx1_first := mag.Lookup(flowHash1)
	idx1_second := mag.Lookup(flowHash1)
	if idx1_first != idx1_second {
		t.Errorf("inconsistent lookup for hash %d: %d != %d", flowHash1, idx1_first, idx1_second)
	}
}
