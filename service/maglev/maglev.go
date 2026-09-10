// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package maglev implements Maglev consistent hashing for service load balancing.
// Maglev provides stable backend selection across rescaling events.
// Reference: Google Maglev paper (Eisenbud et al., NSDI 2016).
//
// Table size M must be a prime number. Default: 251 (or 128 for eBPF BPF_MAP).
package maglev

import "hash/fnv"

const DefaultTableSize = 251

// Table is a Maglev consistent hash lookup table.
type Table struct {
	size     int
	backends []uint32 // backend IDs mapped to slots
}

// Build creates a Maglev lookup table from a set of backend IDs.
// backendIDs must be stable identifiers (not ephemeral).
func Build(backendIDs []uint32, tableSize int) *Table {
	if tableSize == 0 {
		tableSize = DefaultTableSize
	}
	t := &Table{
		size:     tableSize,
		backends: make([]uint32, tableSize),
	}
	if len(backendIDs) == 0 {
		return t
	}

	// Populate preference lists for each backend.
	prefs := make([][]int, len(backendIDs))
	for i, id := range backendIDs {
		prefs[i] = permutation(id, tableSize)
	}

	// Fill the table using the Maglev algorithm.
	table := make([]int, tableSize)
	for i := range table {
		table[i] = -1
	}
	next := make([]int, len(backendIDs))
	filled := 0
	for filled < tableSize {
		for i, id := range backendIDs {
			c := prefs[i][next[i]]
			for table[c] != -1 {
				next[i]++
				c = prefs[i][next[i]]
			}
			table[c] = int(id)
			next[i]++
			filled++
			if filled == tableSize {
				break
			}
		}
	}

	for i, v := range table {
		t.backends[i] = uint32(v)
	}
	return t
}

// Lookup returns the backend ID for the given 4-tuple hash key.
func (t *Table) Lookup(srcIP, dstIP uint32, srcPort, dstPort uint16, proto uint8) uint32 {
	h := fnv.New32a()
	buf := [13]byte{}
	buf[0] = byte(srcIP >> 24)
	buf[1] = byte(srcIP >> 16)
	buf[2] = byte(srcIP >> 8)
	buf[3] = byte(srcIP)
	buf[4] = byte(dstIP >> 24)
	buf[5] = byte(dstIP >> 16)
	buf[6] = byte(dstIP >> 8)
	buf[7] = byte(dstIP)
	buf[8] = byte(srcPort >> 8)
	buf[9] = byte(srcPort)
	buf[10] = byte(dstPort >> 8)
	buf[11] = byte(dstPort)
	buf[12] = proto
	h.Write(buf[:])
	slot := h.Sum32() % uint32(t.size)
	return t.backends[slot]
}

// permutation generates the Maglev permutation for a backend ID.
func permutation(id uint32, m int) []int {
	h1 := fnvHash(id, 0) % uint32(m)
	h2 := fnvHash(id, 1)%(uint32(m)-1) + 1
	perm := make([]int, m)
	for i := range perm {
		perm[i] = int((h1 + uint32(i)*h2) % uint32(m))
	}
	return perm
}

func fnvHash(id, seed uint32) uint32 {
	h := fnv.New32a()
	b := [8]byte{
		byte(id >> 24), byte(id >> 16), byte(id >> 8), byte(id),
		byte(seed >> 24), byte(seed >> 16), byte(seed >> 8), byte(seed),
	}
	h.Write(b[:])
	return h.Sum32()
}
