// Package ipam implements the StraitGateway cluster-pool IPAM backend.
//
// cluster-pool IPAM allocates pod IPs from per-node CIDRs. The CIDRs are
// sourced from the Kubernetes node's spec.podCIDRs field. No RFC1918
// assumption is made — any valid IPv4 or IPv6 CIDR is supported.
package ipam

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

// IPVersion represents IP family type.
type IPVersion int

const (
	IPv4 IPVersion = 4
	IPv6 IPVersion = 6
)

// Allocation contains the allocated IPs and gateways for a pod.
type Allocation struct {
	IPv4     string `json:"ipv4,omitempty"`     // e.g. "10.244.1.5/24"
	IPv6     string `json:"ipv6,omitempty"`     // e.g. "fd00:10:244::5/64"
	Gateway4 net.IP `json:"gateway4,omitempty"` // e.g. 10.244.1.1
	Gateway6 net.IP `json:"gateway6,omitempty"` // e.g. fd00:10:244::1
}

// Allocator manages IP allocation for both IPv4 and IPv6 pod CIDRs.
//
// Concurrency & lifecycle notes (Finding #6):
// In the standard Kubernetes CNI execution model, each CNI ADD, DEL, or CHECK invocation
// executes as an independent short-lived OS process. As such, Allocator instances are not
// shared across CNI processes in memory. Cross-process synchronization and state persistence
// are managed exclusively by Store using flock-based file locking and RestoreFromStore.
// Within a single process or long-running daemon (e.g. StraitD), Allocator is safe for
// concurrent access via its internal mutex.
type Allocator struct {
	mu                 sync.Mutex
	v4Pools            []*v4Pool
	v6Pools            []*v6Pool
	gateway4           net.IP
	gateway6           net.IP
	isDual             bool
	ExhaustionEventsV4 int64
	ExhaustionEventsV6 int64
}

// v4Pool manages an IPv4 CIDR block with a bitmask.
type v4Pool struct {
	network *net.IPNet
	base    uint32 // network address as uint32
	size    uint32 // number of host addresses
	bitmap  []uint64
}

// v6Pool manages an IPv6 CIDR block.
// To handle huge IPv6 subnets (/64 etc.) without petabytes of memory,
// we manage the lowest 65,536 host offsets using a compact bitmap.
type v6Pool struct {
	network *net.IPNet
	base    net.IP // 16-byte base
	size    uint32 // manageable host range (up to 65536)
	bitmap  []uint64
}

const maxV6HostSlots = 65536

// NewAllocator creates an Allocator for the given CIDRs.
// cidrs can contain IPv4 CIDRs (e.g. "10.244.0.0/24"), IPv6 CIDRs (e.g. "fd00:10:244::/64"), or both.
func NewAllocator(cidrs []string) (*Allocator, error) {
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("ipam: at least one CIDR is required")
	}

	var v4Pools []*v4Pool
	var v6Pools []*v6Pool

	for _, cidr := range cidrs {
		ip, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("ipam: invalid CIDR %q: %w", cidr, err)
		}

		if ip4 := ip.To4(); ip4 != nil {
			ones, bits := network.Mask.Size()
			hostBits := uint32(bits - ones)
			if hostBits > 30 {
				hostBits = 30 // cap to prevent excessive memory
			}
			size := uint32(1) << hostBits
			if size < 4 {
				return nil, fmt.Errorf("ipam: IPv4 CIDR %q too small (need at least /30)", cidr)
			}

			base := binary.BigEndian.Uint32(ip4)
			bitmapLen := (size + 63) / 64
			pool := &v4Pool{
				network: network,
				base:    base,
				size:    size,
				bitmap:  make([]uint64, bitmapLen),
			}
			// Mark network address (0) and broadcast (size-1) as allocated.
			pool.markAllocated(0)
			pool.markAllocated(size - 1)
			v4Pools = append(v4Pools, pool)
		} else if ip.To16() != nil {
			ones, _ := network.Mask.Size()
			if ones > 126 {
				return nil, fmt.Errorf("ipam: IPv6 CIDR %q too small (need at least /126)", cidr)
			}

			size := uint32(maxV6HostSlots)
			hostBits := uint32(128 - ones)
			if hostBits < 16 {
				size = uint32(1) << hostBits
			}

			base := make(net.IP, 16)
			copy(base, network.IP.To16())

			bitmapLen := (size + 63) / 64
			pool := &v6Pool{
				network: network,
				base:    base,
				size:    size,
				bitmap:  make([]uint64, bitmapLen),
			}
			// Mark subnet router anycast address (0) as allocated.
			pool.markAllocated(0)
			v6Pools = append(v6Pools, pool)
		}
	}

	if len(v4Pools) == 0 && len(v6Pools) == 0 {
		return nil, fmt.Errorf("ipam: no valid IPv4 or IPv6 pools created")
	}

	alloc := &Allocator{
		v4Pools: v4Pools,
		v6Pools: v6Pools,
		isDual:  len(v4Pools) > 0 && len(v6Pools) > 0,
	}

	// Setup gateways
	if len(v4Pools) > 0 {
		p := v4Pools[0]
		gwRaw := p.base + 1
		gw := make(net.IP, 4)
		binary.BigEndian.PutUint32(gw, gwRaw)
		p.markAllocated(1)
		alloc.gateway4 = gw
	}
	if len(v6Pools) > 0 {
		p := v6Pools[0]
		gw := make(net.IP, 16)
		copy(gw, p.base)
		gw[15] ^= 1 // offset 1
		p.markAllocated(1)
		alloc.gateway6 = gw
	}

	return alloc, nil
}

// Allocate returns the primary IP address (IPv4 if available, otherwise IPv6) as a CIDR string.
// Maintained for CNI backward compatibility.
func (a *Allocator) Allocate() (string, error) {
	alloc, err := a.AllocateDualStack()
	if err != nil {
		return "", err
	}
	if alloc.IPv4 != "" {
		return alloc.IPv4, nil
	}
	return alloc.IPv6, nil
}

// AllocateDualStack allocates both IPv4 and IPv6 addresses if configured.
func (a *Allocator) AllocateDualStack() (*Allocation, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var res Allocation
	var err error

	if len(a.v4Pools) > 0 {
		res.IPv4, err = a.allocateV4Locked()
		if err != nil {
			return nil, err
		}
		res.Gateway4 = a.gateway4
	}

	if len(a.v6Pools) > 0 {
		res.IPv6, err = a.allocateV6Locked()
		if err != nil {
			// Rollback v4 allocation if v6 failed
			if res.IPv4 != "" {
				_ = a.releaseV4Locked(res.IPv4)
			}
			return nil, err
		}
		res.Gateway6 = a.gateway6
	}

	return &res, nil
}

// AllocateV4 allocates an IPv4 address.
func (a *Allocator) AllocateV4() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.allocateV4Locked()
}

// AllocateV6 allocates an IPv6 address.
func (a *Allocator) AllocateV6() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.allocateV6Locked()
}

func (a *Allocator) allocateV4Locked() (string, error) {
	for _, p := range a.v4Pools {
		for i := uint32(1); i < p.size-1; i++ {
			if !p.isAllocated(i) {
				p.markAllocated(i)
				raw := p.base + i
				ip := make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, raw)
				ones, _ := p.network.Mask.Size()
				return fmt.Sprintf("%s/%d", ip.String(), ones), nil
			}
		}
	}
	a.ExhaustionEventsV4++
	return "", fmt.Errorf("ipam: no IPv4 addresses available in any pool")
}

func (a *Allocator) allocateV6Locked() (string, error) {
	for _, p := range a.v6Pools {
		for i := uint32(1); i < p.size; i++ {
			if !p.isAllocated(i) {
				p.markAllocated(i)
				ip := make(net.IP, 16)
				copy(ip, p.base)
				// Add offset to the lowest 4 bytes of IPv6 address
				val := binary.BigEndian.Uint32(ip[12:16]) + i
				binary.BigEndian.PutUint32(ip[12:16], val)
				ones, _ := p.network.Mask.Size()
				return fmt.Sprintf("%s/%d", ip.String(), ones), nil
			}
		}
	}
	a.ExhaustionEventsV6++
	return "", fmt.Errorf("ipam: no IPv6 addresses available in any pool")
}

// Release marks an IP address (IPv4 or IPv6) as available for re-use.
func (a *Allocator) Release(cidr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		// Try parsing as bare IP
		ip = net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("ipam.Release: parse %q: %w", cidr, err)
		}
	}

	if ip4 := ip.To4(); ip4 != nil {
		return a.releaseV4Locked(cidr)
	}
	return a.releaseV6Locked(cidr)
}

func (a *Allocator) releaseV4Locked(cidr string) error {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		ip = net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("ipam.Release: invalid IP: %q", cidr)
		}
	}
	raw := binary.BigEndian.Uint32(ip.To4())

	for _, p := range a.v4Pools {
		if raw <= p.base || raw >= p.base+p.size-1 {
			continue
		}
		offset := raw - p.base
		p.clearAllocated(offset)
		return nil
	}
	return fmt.Errorf("ipam.Release: %q is not in any managed IPv4 pool", cidr)
}

func (a *Allocator) releaseV6Locked(cidr string) error {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		ip = net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("ipam.Release: invalid IP: %q", cidr)
		}
	}
	ip16 := ip.To16()

	for _, p := range a.v6Pools {
		if !p.network.Contains(ip16) {
			continue
		}
		baseVal := binary.BigEndian.Uint32(p.base[12:16])
		ipVal := binary.BigEndian.Uint32(ip16[12:16])
		if ipVal < baseVal || ipVal >= baseVal+p.size {
			continue
		}
		offset := ipVal - baseVal
		p.clearAllocated(offset)
		return nil
	}
	return fmt.Errorf("ipam.Release: %q is not in any managed IPv6 pool", cidr)
}

// Reserve marks a specific IP address as allocated/reserved.
func (a *Allocator) Reserve(cidr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		ip = net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("ipam.Reserve: invalid IP: %q", cidr)
		}
	}

	if ip4 := ip.To4(); ip4 != nil {
		raw := binary.BigEndian.Uint32(ip4)
		for _, p := range a.v4Pools {
			if raw >= p.base && raw < p.base+p.size {
				offset := raw - p.base
				p.markAllocated(offset)
				return nil
			}
		}
		return fmt.Errorf("ipam.Reserve: %q not in IPv4 pools", cidr)
	}

	ip16 := ip.To16()
	for _, p := range a.v6Pools {
		if p.network.Contains(ip16) {
			baseVal := binary.BigEndian.Uint32(p.base[12:16])
			ipVal := binary.BigEndian.Uint32(ip16[12:16])
			if ipVal >= baseVal && ipVal < baseVal+p.size {
				offset := ipVal - baseVal
				p.markAllocated(offset)
				return nil
			}
		}
	}
	return fmt.Errorf("ipam.Reserve: %q not in IPv6 pools", cidr)
}

// RestoreFromStore reconciles all entries loaded from the persistent store into allocator bitmaps.
func (a *Allocator) RestoreFromStore(store *Store) error {
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load store for IPAM restoration: %w", err)
	}

	for _, alloc := range state.Allocations {
		if alloc.IPv4 != "" {
			_ = a.Reserve(alloc.IPv4)
		}
		if alloc.IPv6 != "" {
			_ = a.Reserve(alloc.IPv6)
		}
	}
	return nil
}

// Gateway returns the gateway IP for the first pool (backward compatibility).
func (a *Allocator) Gateway() net.IP { return a.gateway4 }

// Gateway4 returns the IPv4 gateway.
func (a *Allocator) Gateway4() net.IP { return a.gateway4 }

// Gateway6 returns the IPv6 gateway.
func (a *Allocator) Gateway6() net.IP { return a.gateway6 }

// IsDualStack returns whether allocator is managing both IPv4 and IPv6.
func (a *Allocator) IsDualStack() bool { return a.isDual }

func (p *v4Pool) isAllocated(offset uint32) bool {
	word := offset / 64
	bit := offset % 64
	return p.bitmap[word]&(1<<bit) != 0
}

func (p *v4Pool) markAllocated(offset uint32) {
	word := offset / 64
	bit := offset % 64
	p.bitmap[word] |= 1 << bit
}

func (p *v4Pool) clearAllocated(offset uint32) {
	word := offset / 64
	bit := offset % 64
	p.bitmap[word] &^= 1 << bit
}

func (p *v6Pool) isAllocated(offset uint32) bool {
	word := offset / 64
	bit := offset % 64
	return p.bitmap[word]&(1<<bit) != 0
}

func (p *v6Pool) markAllocated(offset uint32) {
	word := offset / 64
	bit := offset % 64
	p.bitmap[word] |= 1 << bit
}

func (p *v6Pool) clearAllocated(offset uint32) {
	word := offset / 64
	bit := offset % 64
	p.bitmap[word] &^= 1 << bit
}
