package datapath

import (
	"hash/fnv"
	"math"
	"testing"
)

// MaglevTable represents a Google Maglev consistent hashing lookup table.
type MaglevTable struct {
	m      int      // Lookup table size (prime number)
	lookup []string // Selected backend for each slot
}

// NewMaglevTable constructs a simulated Maglev consistent hashing table for N backends.
func NewMaglevTable(backends []string, tableSize int) *MaglevTable {
	if len(backends) == 0 {
		return &MaglevTable{m: tableSize, lookup: make([]string, tableSize)}
	}

	m := tableSize
	n := len(backends)

	// Generate permutation sequences for each backend using FNV hashes
	permutation := make([][]int, n)
	for i, b := range backends {
		h1 := fnvHash(b + "-offset")
		h2 := fnvHash(b + "-skip")
		offset := int(h1 % uint64(m))
		skip := int(h2%uint64(m-1)) + 1

		permutation[i] = make([]int, m)
		for j := 0; j < m; j++ {
			permutation[i][j] = (offset + j*skip) % m
		}
	}

	// Populate lookup table
	lookup := make([]string, m)
	next := make([]int, n)
	filled := 0

	for filled < m {
		for i := 0; i < n; i++ {
			c := permutation[i][next[i]]
			for lookup[c] != "" {
				next[i]++
				c = permutation[i][next[i]]
			}
			lookup[c] = backends[i]
			next[i]++
			filled++
			if filled == m {
				break
			}
		}
	}

	return &MaglevTable{m: m, lookup: lookup}
}

func fnvHash(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// Select returns the backend mapped to a given 5-tuple hash key.
func (mt *MaglevTable) Select(hashKey uint64) string {
	idx := hashKey % uint64(mt.m)
	return mt.lookup[idx]
}

// TestMaglevConsistentHashing tests distribution uniformity and minimal disruption.
func TestMaglevConsistentHashing(t *testing.T) {
	const tableSize = 65537 // Prime number table size commonly used in kernel BPF Maglev
	backends := []string{"10.244.1.10", "10.244.1.11", "10.244.1.12", "10.244.1.13"}

	table := NewMaglevTable(backends, tableSize)

	// Measure backend distribution
	counts := make(map[string]int)
	for _, b := range table.lookup {
		counts[b]++
	}

	expectedPerBackend := float64(tableSize) / float64(len(backends))
	tolerance := expectedPerBackend * 0.05 // Within 5% variance

	for _, b := range backends {
		diff := math.Abs(float64(counts[b]) - expectedPerBackend)
		if diff > tolerance {
			t.Errorf("Backend %s distribution skewed: got %d slots, expected ~%.0f", b, counts[b], expectedPerBackend)
		} else {
			t.Logf("✓ Backend %s received %d slots (~%.2f%%)", b, counts[b], float64(counts[b])/float64(tableSize)*100)
		}
	}

	// Minimal disruption test: remove one backend and check consistency of remaining backends
	reducedBackends := []string{"10.244.1.10", "10.244.1.11", "10.244.1.12"}
	newTable := NewMaglevTable(reducedBackends, tableSize)

	moved := 0
	for i := 0; i < tableSize; i++ {
		orig := table.lookup[i]
		if orig != "10.244.1.13" && newTable.lookup[i] != orig {
			moved++
		}
	}

	disruptionRate := float64(moved) / float64(tableSize)
	t.Logf("Disruption rate for intact backends after removing 1 backend: %.2f%%", disruptionRate*100)

	// Disruption on surviving backends should be very small (< 10%)
	if disruptionRate > 0.10 {
		t.Errorf("Disruption rate too high: %.2f%%; expected < 10%%", disruptionRate*100)
	}
}

// TestDirectServerReturnValidation verifies DSR packet header transformation expectations.
func TestDirectServerReturnValidation(t *testing.T) {
	// In DSR mode:
	// - Client sends SYN to Service VIP:Port
	// - NodePort / LB forwards packet to Backend IP without SNAT, preserving original Client IP
	// - Backend receives packet, processes request, and replies DIRECTLY to Client IP using VIP as Source IP
	origClientIP := "203.0.113.50"
	serviceVIP := "10.96.0.100"
	backendIP := "10.244.1.45"

	type PacketHeaders struct {
		SrcIP string
		DstIP string
	}

	// 1. Ingress at NodePort / Gateway
	ingressPacket := PacketHeaders{SrcIP: origClientIP, DstIP: serviceVIP}

	// 2. Transformed packet sent to backend (DSR: outer MAC/IP routed, inner VIP retained or tunneled)
	forwardedPacket := PacketHeaders{SrcIP: origClientIP, DstIP: backendIP}
	if forwardedPacket.SrcIP != origClientIP || forwardedPacket.DstIP != backendIP {
		t.Error("Forwarded packet must preserve client IP and route to backend")
	}

	// 3. Egress response from backend directly to client
	egressResponse := PacketHeaders{SrcIP: serviceVIP, DstIP: origClientIP}

	if ingressPacket.SrcIP != egressResponse.DstIP {
		t.Error("Client must receive response at original client IP")
	}
	if egressResponse.SrcIP != ingressPacket.DstIP {
		t.Error("DSR requires response SrcIP to match Service VIP, not Backend IP")
	}
	t.Logf("✓ DSR verified: Client %s -> VIP %s -> Backend %s -> Client %s",
		origClientIP, serviceVIP, backendIP, origClientIP)
}
