// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	"github.com/msaeedb40/straitgateway/service/maglev"
)

func TestMaglevTableBuild(t *testing.T) {
	backends := []uint32{1, 2, 3, 4, 5}
	table := maglev.Build(backends, maglev.DefaultTableSize)

	if table == nil {
		t.Fatal("maglev.Build returned nil")
	}

	// Verify all slots are filled with valid backend IDs.
	for i := 0; i < maglev.DefaultTableSize; i++ {
		found := false
		result := table.Lookup(uint32(i), 0, 0, 0, 6)
		for _, be := range backends {
			if result == be {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("slot %d has invalid backend %d", i, result)
		}
	}
}

func TestMaglevConsistency(t *testing.T) {
	backends := []uint32{1, 2, 3}
	table := maglev.Build(backends, maglev.DefaultTableSize)

	// Same input → same output (deterministic).
	r1 := table.Lookup(0x0A000001, 0x0A000002, 8080, 80, 6)
	r2 := table.Lookup(0x0A000001, 0x0A000002, 8080, 80, 6)
	if r1 != r2 {
		t.Errorf("non-deterministic: got %d then %d", r1, r2)
	}
}

func TestMaglevEmptyBackends(t *testing.T) {
	table := maglev.Build(nil, maglev.DefaultTableSize)
	if table == nil {
		t.Fatal("maglev.Build(nil) returned nil")
	}
}
