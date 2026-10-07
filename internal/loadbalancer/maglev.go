package loadbalancer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// DefaultMaglevTableSize is a large prime number for Maglev hashing distribution
const DefaultMaglevTableSize uint64 = 65537

// Maglev generates consistent hashing lookup tables for backend load balancing
type Maglev struct {
	tableSize uint64
	backends  []string
	lookup    []int
}

// NewMaglev creates a Maglev hash generator for given backend names
func NewMaglev(backends []string, tableSize uint64) (*Maglev, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("backend list cannot be empty")
	}
	if tableSize == 0 {
		tableSize = DefaultMaglevTableSize
	}

	m := &Maglev{
		tableSize: tableSize,
		backends:  backends,
		lookup:    make([]int, tableSize),
	}
	m.generate()
	return m, nil
}

func (m *Maglev) generate() {
	n := len(m.backends)
	permutations := make([][]uint64, n)

	for i, b := range m.backends {
		offset, skip := m.hash(b)
		permutations[i] = make([]uint64, m.tableSize)
		for j := uint64(0); j < m.tableSize; j++ {
			permutations[i][j] = (offset + j*skip) % m.tableSize
		}
	}

	for i := range m.lookup {
		m.lookup[i] = -1
	}

	next := make([]int, n)
	filled := uint64(0)

	for {
		for i := 0; i < n; i++ {
			c := permutations[i][next[i]]
			for m.lookup[c] >= 0 {
				next[i]++
				c = permutations[i][next[i]]
			}
			m.lookup[c] = i
			next[i]++
			filled++
			if filled == m.tableSize {
				return
			}
		}
	}
}

func (m *Maglev) hash(backend string) (offset, skip uint64) {
	h := sha256.Sum256([]byte(backend))
	h1 := binary.BigEndian.Uint64(h[:8])
	h2 := binary.BigEndian.Uint64(h[8:16])

	offset = h1 % m.tableSize
	skip = (h2 % (m.tableSize - 1)) + 1
	return
}

// Lookup returns the selected backend index for a packet 5-tuple hash
func (m *Maglev) Lookup(flowHash uint64) int {
	pos := flowHash % m.tableSize
	return m.lookup[pos]
}

// LookupTable returns the precomputed table
func (m *Maglev) LookupTable() []int {
	return m.lookup
}
