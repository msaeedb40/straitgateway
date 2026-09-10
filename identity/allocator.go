// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package identity implements the 32-bit security identity allocator.
//
// Identities are assigned to pods based on a hash of their sorted label set.
// Pods with identical labels share the same identity. This enables efficient
// policy enforcement in BPF maps (identity_map) without per-pod entries.
//
// Architecture:
//   - Identity 0 is reserved (unknown/world).
//   - Identities 1–255 are reserved for system identities (kube-dns, etc.).
//   - Identities 256+ are dynamically allocated via the controller.
//   - Identity GC runs periodically to reclaim unused allocations.
package identity

import (
	"hash/fnv"
	"sort"
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

const (
	// MinDynamicIdentity is the first allocatable identity ID.
	MinDynamicIdentity sgtypes.Identity = 256
	// IdentityWorld represents unknown/unidentified traffic.
	IdentityWorld sgtypes.Identity = 0
	// IdentityKubeDNS is the reserved identity for kube-dns/CoreDNS.
	IdentityKubeDNS sgtypes.Identity = 2
	// IdentityHealth is the reserved identity for health checks.
	IdentityHealth sgtypes.Identity = 4
)

// Allocator manages the allocation and deallocation of security identities.
type Allocator struct {
	mu     sync.RWMutex
	// byLabelsHash maps label hash → identity.
	byLabelsHash map[uint64]sgtypes.Identity
	// byID maps identity → label hash + metadata.
	byID   map[sgtypes.Identity]*IdentityEntry
	// nextID is the next available identity to allocate.
	nextID sgtypes.Identity
	// refCount tracks how many pods share each identity.
	refCount map[sgtypes.Identity]int
}

// IdentityEntry stores metadata for an allocated identity.
type IdentityEntry struct {
	Identity   sgtypes.Identity
	LabelsHash uint64
	Labels     map[string]string
	SegmentID  sgtypes.SegmentID
}

// NewAllocator creates a new identity allocator.
func NewAllocator() *Allocator {
	return &Allocator{
		byLabelsHash: make(map[uint64]sgtypes.Identity),
		byID:         make(map[sgtypes.Identity]*IdentityEntry),
		refCount:     make(map[sgtypes.Identity]int),
		nextID:       MinDynamicIdentity,
	}
}

// Allocate returns an identity for the given label set.
// If an identity already exists for these labels, it is reused and ref-counted.
func (a *Allocator) Allocate(labels map[string]string, segmentID sgtypes.SegmentID) *IdentityEntry {
	hash := HashLabels(labels)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if identity already exists for this label set.
	if id, ok := a.byLabelsHash[hash]; ok {
		a.refCount[id]++
		return a.byID[id]
	}

	// Allocate new identity.
	id := a.nextID
	a.nextID++

	entry := &IdentityEntry{
		Identity:   id,
		LabelsHash: hash,
		Labels:     labels,
		SegmentID:  segmentID,
	}

	a.byLabelsHash[hash] = id
	a.byID[id] = entry
	a.refCount[id] = 1

	return entry
}

// Release decrements the ref count for an identity. Returns true if the identity was freed.
func (a *Allocator) Release(id sgtypes.Identity) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.refCount[id]--
	if a.refCount[id] <= 0 {
		// Identity is no longer in use — mark for GC.
		entry, ok := a.byID[id]
		if ok {
			delete(a.byLabelsHash, entry.LabelsHash)
			delete(a.byID, id)
		}
		delete(a.refCount, id)
		return true
	}
	return false
}

// Lookup returns the identity entry for the given ID, or nil if not found.
func (a *Allocator) Lookup(id sgtypes.Identity) *IdentityEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.byID[id]
}

// LookupByLabels returns the identity for the given label set, or 0 if not found.
func (a *Allocator) LookupByLabels(labels map[string]string) sgtypes.Identity {
	hash := HashLabels(labels)
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.byLabelsHash[hash]
}

// AllEntries returns all currently allocated identity entries.
func (a *Allocator) AllEntries() []*IdentityEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	entries := make([]*IdentityEntry, 0, len(a.byID))
	for _, e := range a.byID {
		entries = append(entries, e)
	}
	return entries
}

// HashLabels computes an FNV-1a hash of the sorted label key=value pairs.
func HashLabels(labels map[string]string) uint64 {
	// Sort keys for deterministic hashing.
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := fnv.New64a()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte("="))
		h.Write([]byte(labels[k]))
		h.Write([]byte(","))
	}
	return h.Sum64()
}
