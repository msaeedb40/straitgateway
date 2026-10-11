package routing

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/pkg/types"
)

// RoutePrecedence defines administrative preference (lower value = higher priority)
type RoutePrecedence uint32

const (
	PrecedenceLocalPod     RoutePrecedence = 10
	PrecedenceClusterRoute RoutePrecedence = 20
	PrecedenceStaticRoute  RoutePrecedence = 30
	PrecedenceBGPRoute     RoutePrecedence = 40
	PrecedenceDefaultRoute RoutePrecedence = 100
)

// RouteEntry represents an active L3 forwarding route
type RouteEntry struct {
	Destination net.IPNet
	VRFID       uint32
	NextHopID   uint32
	NextHopIP   net.IP
	Interface   string
	Precedence  RoutePrecedence
	Metric      uint32
	Communities []string
}

// VRFTable manages routes for an isolated routing domain
type VRFTable struct {
	ID          uint32
	Name        string
	Routes      map[string]*RouteEntry // key: CIDR string
	Description string
}

// FIBEngine manages multi-VRF forwarding information base
type FIBEngine struct {
	mu     sync.RWMutex
	vrfs   map[uint32]*VRFTable
	routes map[string]*RouteEntry // Global lookup key: vrfID:cidr
}

// NewFIBEngine creates a new Forwarding Information Base engine
func NewFIBEngine() *FIBEngine {
	fib := &FIBEngine{
		vrfs:   make(map[uint32]*VRFTable),
		routes: make(map[string]*RouteEntry),
	}
	// Default VRF 0
	fib.vrfs[0] = &VRFTable{
		ID:     0,
		Name:   "default",
		Routes: make(map[string]*RouteEntry),
	}
	return fib
}

// AddVRF creates an isolated routing domain
func (f *FIBEngine) AddVRF(id uint32, name, description string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.vrfs[id]; !exists {
		f.vrfs[id] = &VRFTable{
			ID:          id,
			Name:        name,
			Description: description,
			Routes:      make(map[string]*RouteEntry),
		}
	}
}

// InsertRoute adds or updates a route with precedence checking
func (f *FIBEngine) InsertRoute(entry *RouteEntry) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	vrf, exists := f.vrfs[entry.VRFID]
	if !exists {
		return false, fmt.Errorf("VRF %d does not exist", entry.VRFID)
	}

	cidrKey := entry.Destination.String()
	lookupKey := fmt.Sprintf("%d:%s", entry.VRFID, cidrKey)

	// Check existing route precedence
	if existing, found := vrf.Routes[cidrKey]; found {
		if entry.Precedence > existing.Precedence {
			// Inferior route; do not replace
			return false, nil
		}
		if entry.Precedence == existing.Precedence && entry.Metric > existing.Metric {
			// Higher metric; do not replace
			return false, nil
		}
	}

	vrf.Routes[cidrKey] = entry
	f.routes[lookupKey] = entry
	return true, nil
}

// DeleteRoute removes a route from the specified VRF
func (f *FIBEngine) DeleteRoute(vrfID uint32, dst net.IPNet) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	vrf, exists := f.vrfs[vrfID]
	if !exists {
		return false
	}

	cidrKey := dst.String()
	lookupKey := fmt.Sprintf("%d:%s", vrfID, cidrKey)

	if _, found := vrf.Routes[cidrKey]; found {
		delete(vrf.Routes, cidrKey)
		delete(f.routes, lookupKey)
		return true
	}
	return false
}

// LongestPrefixMatch performs LPM lookup in the given VRF
func (f *FIBEngine) LongestPrefixMatch(vrfID uint32, ip net.IP) (*RouteEntry, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	vrf, exists := f.vrfs[vrfID]
	if !exists {
		return nil, fmt.Errorf("VRF %d not found", vrfID)
	}

	var bestMatch *RouteEntry
	var maxPrefixLen int = -1

	for _, route := range vrf.Routes {
		if route.Destination.Contains(ip) {
			ones, _ := route.Destination.Mask.Size()
			if ones > maxPrefixLen {
				maxPrefixLen = ones
				bestMatch = route
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no route to host %s in VRF %d", ip, vrfID)
	}
	return bestMatch, nil
}

// ToLPMV4Key converts a route destination to eBPF LPM key
func ToLPMV4Key(n net.IPNet) (types.LPMV4Key, error) {
	ip4 := n.IP.To4()
	if ip4 == nil {
		return types.LPMV4Key{}, fmt.Errorf("not an IPv4 address: %v", n.IP)
	}
	ones, _ := n.Mask.Size()
	return types.LPMV4Key{
		PrefixLen: uint32(ones),
		Addr:      binary.BigEndian.Uint32(ip4),
	}, nil
}

// ListRoutes returns all routes configured for a given VRF
func (f *FIBEngine) ListRoutes(vrfID uint32) []*RouteEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()

	vrf, exists := f.vrfs[vrfID]
	if !exists {
		return nil
	}

	result := make([]*RouteEntry, 0, len(vrf.Routes))
	for _, r := range vrf.Routes {
		result = append(result, r)
	}
	return result
}

// ListVRFs returns a list of all active VRF tables
func (f *FIBEngine) ListVRFs() []*VRFTable {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]*VRFTable, 0, len(f.vrfs))
	for _, vrf := range f.vrfs {
		result = append(result, vrf)
	}
	return result
}

// GetVRF retrieves a VRF by ID
func (f *FIBEngine) GetVRF(id uint32) (*VRFTable, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	vrf, exists := f.vrfs[id]
	return vrf, exists
}

// FindRoutesByNextHop returns all routes across all VRFs associated with a given next-hop IP
func (f *FIBEngine) FindRoutesByNextHop(nextHopIP net.IP) []*RouteEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var matching []*RouteEntry
	for _, vrf := range f.vrfs {
		for _, r := range vrf.Routes {
			if r.NextHopIP != nil && r.NextHopIP.Equal(nextHopIP) {
				matching = append(matching, r)
			}
		}
	}
	return matching
}
