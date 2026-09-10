// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	"github.com/msaeedb40/straitgateway/identity"
)

func TestIdentityAllocator(t *testing.T) {
	alloc := identity.NewAllocator()

	labels := map[string]string{"app": "web", "env": "prod"}
	entry := alloc.Allocate(labels, 0)

	if entry.Identity < identity.MinDynamicIdentity {
		t.Errorf("identity %d is below minimum %d", entry.Identity, identity.MinDynamicIdentity)
	}

	// Same labels should return same identity.
	entry2 := alloc.Allocate(labels, 0)
	if entry2.Identity != entry.Identity {
		t.Errorf("expected reuse: got %d, want %d", entry2.Identity, entry.Identity)
	}

	// Different labels should get different identity.
	labels2 := map[string]string{"app": "api", "env": "prod"}
	entry3 := alloc.Allocate(labels2, 0)
	if entry3.Identity == entry.Identity {
		t.Error("different labels got same identity")
	}
}

func TestIdentityRelease(t *testing.T) {
	alloc := identity.NewAllocator()

	labels := map[string]string{"app": "test"}
	e1 := alloc.Allocate(labels, 0)
	alloc.Allocate(labels, 0) // ref count = 2

	freed := alloc.Release(e1.Identity) // ref count = 1
	if freed {
		t.Error("should not be freed yet (ref count = 1)")
	}

	freed = alloc.Release(e1.Identity) // ref count = 0
	if !freed {
		t.Error("should be freed (ref count = 0)")
	}

	// Identity should no longer be found.
	if alloc.Lookup(e1.Identity) != nil {
		t.Error("identity still found after full release")
	}
}

func TestIdentityLabelHashDeterministic(t *testing.T) {
	labels := map[string]string{"z": "1", "a": "2", "m": "3"}
	h1 := identity.HashLabels(labels)
	h2 := identity.HashLabels(labels)
	if h1 != h2 {
		t.Errorf("non-deterministic hash: %d != %d", h1, h2)
	}
}
