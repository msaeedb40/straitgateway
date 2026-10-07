package maps

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/straitgateway/straitgateway/pkg/types"
)

const (
	DefaultBPFFSPath = "/sys/fs/bpf/straitgateway"

	// Map names
	RouteMapName     = "sg_routes"
	NextHopMapName   = "sg_nexthops"
	ServiceMapName   = "sg_services"
	EndpointMapName  = "sg_endpoints"
	ConntrackMapName = "sg_conntrack"
	PolicyMapName    = "sg_policies"
	CounterMapName   = "sg_counters"
)

// MapManager handles life cycle and access to all StraitGateway BPF maps
type MapManager struct {
	mu        sync.RWMutex
	pinPath   string
	routes    *ebpf.Map
	nexthops  *ebpf.Map
	services  *ebpf.Map
	endpoints *ebpf.Map
	conntrack *ebpf.Map
	policies  *ebpf.Map
	counters  *ebpf.Map
}

// Config defines map sizes and pin paths
type Config struct {
	PinPath          string
	MaxRoutes        uint32
	MaxNextHops      uint32
	MaxServices      uint32
	MaxEndpoints     uint32
	MaxConntrack     uint32
	MaxPolicies      uint32
}

// DefaultConfig provides standard sizing
func DefaultConfig() Config {
	return Config{
		PinPath:      DefaultBPFFSPath,
		MaxRoutes:    65536,
		MaxNextHops:  16384,
		MaxServices:  16384,
		MaxEndpoints: 65536,
		MaxConntrack: 262144,
		MaxPolicies:  32768,
	}
}

// NewMapManager initializes or opens pinned BPF maps
func NewMapManager(cfg Config) (*MapManager, error) {
	if cfg.PinPath == "" {
		cfg.PinPath = DefaultBPFFSPath
	}

	mm := &MapManager{
		pinPath: cfg.PinPath,
	}

	// In test environments or when running unprivileged, support memory-backed maps
	if err := mm.initMaps(cfg); err != nil {
		return nil, fmt.Errorf("failed to init BPF maps: %w", err)
	}

	return mm, nil
}

func (m *MapManager) initMaps(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Routes (LPM Trie)
	routeSpec := &ebpf.MapSpec{
		Name:       RouteMapName,
		Type:       ebpf.LPMTrie,
		KeySize:    8, // struct lpm_v4_key (prefixlen 4 + addr 4)
		ValueSize:  4, // nexthop_id (uint32)
		MaxEntries: cfg.MaxRoutes,
		Flags:      1, // BPF_F_NO_PREALLOC
	}
	rMap, err := ebpf.NewMap(routeSpec)
	if err != nil {
		return fmt.Errorf("create route map: %w", err)
	}
	m.routes = rMap

	// 2. NextHops (Hash)
	nhSpec := &ebpf.MapSpec{
		Name:       NextHopMapName,
		Type:       ebpf.Hash,
		KeySize:    4,  // nexthop_id (uint32)
		ValueSize:  20, // struct nexthop_info
		MaxEntries: cfg.MaxNextHops,
	}
	nhMap, err := ebpf.NewMap(nhSpec)
	if err != nil {
		return fmt.Errorf("create nexthop map: %w", err)
	}
	m.nexthops = nhMap

	// 3. Services (Hash)
	svcSpec := &ebpf.MapSpec{
		Name:       ServiceMapName,
		Type:       ebpf.Hash,
		KeySize:    8, // struct service_key
		ValueSize:  4, // endpoint_count or backend_id
		MaxEntries: cfg.MaxServices,
	}
	svcMap, err := ebpf.NewMap(svcSpec)
	if err != nil {
		return fmt.Errorf("create service map: %w", err)
	}
	m.services = svcMap

	// 4. Endpoints (Array / Hash)
	epSpec := &ebpf.MapSpec{
		Name:       EndpointMapName,
		Type:       ebpf.Hash,
		KeySize:    4, // endpoint_id
		ValueSize:  8, // struct service_endpoint
		MaxEntries: cfg.MaxEndpoints,
	}
	epMap, err := ebpf.NewMap(epSpec)
	if err != nil {
		return fmt.Errorf("create endpoint map: %w", err)
	}
	m.endpoints = epMap

	// 5. Conntrack (LRU Hash)
	ctSpec := &ebpf.MapSpec{
		Name:       ConntrackMapName,
		Type:       ebpf.LRUHash,
		KeySize:    16, // struct flow_key
		ValueSize:  40, // struct flow_entry
		MaxEntries: cfg.MaxConntrack,
	}
	ctMap, err := ebpf.NewMap(ctSpec)
	if err != nil {
		return fmt.Errorf("create conntrack map: %w", err)
	}
	m.conntrack = ctMap

	// 6. Policy (Hash)
	polSpec := &ebpf.MapSpec{
		Name:       PolicyMapName,
		Type:       ebpf.Hash,
		KeySize:    12, // struct policy_key
		ValueSize:  4,  // policy_action (uint32)
		MaxEntries: cfg.MaxPolicies,
	}
	polMap, err := ebpf.NewMap(polSpec)
	if err != nil {
		return fmt.Errorf("create policy map: %w", err)
	}
	m.policies = polMap

	// 7. Counters (Per-CPU Array)
	cntSpec := &ebpf.MapSpec{
		Name:       CounterMapName,
		Type:       ebpf.PerCPUArray,
		KeySize:    4, // counter index
		ValueSize:  8, // uint64 count
		MaxEntries: 64,
	}
	cntMap, err := ebpf.NewMap(cntSpec)
	if err != nil {
		return fmt.Errorf("create counter map: %w", err)
	}
	m.counters = cntMap

	return nil
}

// PinAll pins all maps to bpffs if directory exists and is writable
func (m *MapManager) PinAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.pinPath, 0755); err != nil {
		return fmt.Errorf("create bpffs dir %s: %w", m.pinPath, err)
	}

	maps := map[string]*ebpf.Map{
		RouteMapName:     m.routes,
		NextHopMapName:   m.nexthops,
		ServiceMapName:   m.services,
		EndpointMapName:  m.endpoints,
		ConntrackMapName: m.conntrack,
		PolicyMapName:    m.policies,
		CounterMapName:   m.counters,
	}

	for name, bpfMap := range maps {
		target := filepath.Join(m.pinPath, name)
		_ = os.Remove(target) // remove stale pin if present
		if err := bpfMap.Pin(target); err != nil {
			return fmt.Errorf("failed to pin %s: %w", name, err)
		}
	}
	return nil
}

// AddRoute inserts or updates an LPM route pointing to a next-hop ID
func (m *MapManager) AddRoute(key types.LPMV4Key, nextHopID uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.routes.Put(key, nextHopID)
}

// DeleteRoute removes an LPM route
func (m *MapManager) DeleteRoute(key types.LPMV4Key) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.routes.Delete(key)
}

// LookupRoute retrieves the next-hop ID for an LPM route key
func (m *MapManager) LookupRoute(key types.LPMV4Key) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var nextHopID uint32
	if err := m.routes.Lookup(key, &nextHopID); err != nil {
		return 0, err
	}
	return nextHopID, nil
}

// AddNextHop stores next-hop forwarding parameters
func (m *MapManager) AddNextHop(id uint32, nh types.NextHopInfo) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nexthops.Put(id, nh)
}

// DeleteNextHop removes a next-hop entry
func (m *MapManager) DeleteNextHop(id uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nexthops.Delete(id)
}

// SetPolicy sets allow/deny/audit action for source and destination identities
func (m *MapManager) SetPolicy(key types.PolicyKey, action types.PolicyAction) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.policies.Put(key, uint32(action))
}

// DeletePolicy removes a policy rule
func (m *MapManager) DeletePolicy(key types.PolicyKey) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.policies.Delete(key)
}

// Close releases all loaded BPF map handles
func (m *MapManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	maps := []*ebpf.Map{
		m.routes, m.nexthops, m.services, m.endpoints,
		m.conntrack, m.policies, m.counters,
	}
	for _, bpfMap := range maps {
		if bpfMap != nil {
			if err := bpfMap.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing maps: %v", errs)
	}
	return nil
}
