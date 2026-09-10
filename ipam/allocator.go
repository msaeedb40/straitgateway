// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package ipam implements the per-node IP address allocator for pod networking.
//
// Architecture:
//   - Each node receives a /24 (or larger) CIDR from the cluster podCIDR.
//   - CIDR discovery reads node.spec.podCIDR via the Kubernetes API.
//   - Addresses are allocated from a bitmap to avoid fragmentation.
//   - The allocator runs inside straitgatewayd and serves the CNI plugin via gRPC.
package ipam

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
)

// Allocator manages IP address allocation from a per-node CIDR.
type Allocator struct {
	mu      sync.Mutex
	cidr    netip.Prefix
	gateway netip.Addr // first address in CIDR (gateway)
	bitmap  []bool     // true = allocated
	size    int        // total number of allocatable addresses
}

// New creates a new IPAM Allocator from a CIDR prefix.
// The first address is reserved as the gateway.
func New(cidr netip.Prefix) (*Allocator, error) {
	if !cidr.IsValid() {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	bits := cidr.Bits()
	var size int
	if cidr.Addr().Is4() {
		size = 1 << (32 - bits)
	} else {
		size = 1 << (128 - bits)
		if size > 65536 {
			size = 65536 // cap IPv6 allocations
		}
	}

	// Reserve network address (index 0) and broadcast (last).
	// Reserve index 1 as gateway.
	a := &Allocator{
		cidr:    cidr,
		bitmap:  make([]bool, size),
		size:    size,
		gateway: cidr.Addr().Next(), // .1 is gateway
	}
	a.bitmap[0] = true           // network address
	a.bitmap[1] = true           // gateway
	if size > 2 {
		a.bitmap[size-1] = true  // broadcast
	}

	return a, nil
}

// Allocate assigns the next available IP address.
func (a *Allocator) Allocate() (netip.Addr, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i := 2; i < a.size-1; i++ {
		if !a.bitmap[i] {
			a.bitmap[i] = true
			addr := offsetAddr(a.cidr.Addr(), i)
			return addr, nil
		}
	}
	return netip.Addr{}, fmt.Errorf("CIDR %s exhausted", a.cidr)
}

// AllocateSpecific assigns a specific IP address.
func (a *Allocator) AllocateSpecific(addr netip.Addr) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	offset := addrOffset(a.cidr.Addr(), addr)
	if offset < 0 || offset >= a.size {
		return fmt.Errorf("address %s is outside CIDR %s", addr, a.cidr)
	}
	if a.bitmap[offset] {
		return fmt.Errorf("address %s is already allocated", addr)
	}
	a.bitmap[offset] = true
	return nil
}

// Release releases a previously allocated IP address.
func (a *Allocator) Release(addr netip.Addr) {
	a.mu.Lock()
	defer a.mu.Unlock()

	offset := addrOffset(a.cidr.Addr(), addr)
	if offset >= 2 && offset < a.size-1 {
		a.bitmap[offset] = false
	}
}

// Gateway returns the gateway address for this CIDR.
func (a *Allocator) Gateway() netip.Addr {
	return a.gateway
}

// CIDR returns the managed CIDR prefix.
func (a *Allocator) CIDR() netip.Prefix {
	return a.cidr
}

// Available returns the number of unallocated addresses.
func (a *Allocator) Available() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for i := 2; i < a.size-1; i++ {
		if !a.bitmap[i] {
			count++
		}
	}
	return count
}

// offsetAddr returns the address at offset n from base.
func offsetAddr(base netip.Addr, n int) netip.Addr {
	b := base.As4()
	ip := net.IPv4(b[0], b[1], b[2], b[3]).To4()
	val := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	val += uint32(n)
	return netip.AddrFrom4([4]byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)})
}

// addrOffset returns the offset of addr from base.
func addrOffset(base, addr netip.Addr) int {
	b := base.As4()
	a := addr.As4()
	bVal := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	aVal := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	return int(aVal - bVal)
}
