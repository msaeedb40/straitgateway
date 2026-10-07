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

	// Map names conforming to StraitGateway BPF specification
	RouteMapName        = "sg_routes"
	NextHopMapName      = "sg_nexthops"
	ServiceMapName      = "sg_services"
	EndpointMapName     = "sg_endpoints"
	ConntrackMapName    = "sg_conntrack"
	PolicyMapName       = "sg_policies"
	CounterMapName      = "sg_counters"
	IdentityMapName     = "sg_identities"
	NATMapName          = "sg_nat"
	VRFMapName          = "sg_vrf"
	TunnelMapName       = "sg_tunnels"
	NeighborMapName     = "sg_neighbors"
	EventRingBufferName = "sg_events"
)

// MapManager handles life cycle and access to all StraitGateway BPF maps
type MapManager struct {
	mu         sync.RWMutex
	pinPath    string
	routes     *ebpf.Map
	nexthops   *ebpf.Map
	services   *ebpf.Map
	endpoints  *ebpf.Map
	conntrack  *ebpf.Map
	policies   *ebpf.Map
	counters   *ebpf.Map
	identities *ebpf.Map
	nat        *ebpf.Map
	vrf        *ebpf.Map
	tunnels    *ebpf.Map
	neighbors  *ebpf.Map
	events     *ebpf.Map
}

// Config defines map sizes and pin paths
type Config struct {
	PinPath       string
	MaxRoutes     uint32
	MaxNextHops   uint32
	MaxServices   uint32
	MaxEndpoints  uint32
	MaxConntrack  uint32
	MaxPolicies   uint32
	MaxIdentities uint32
	MaxNAT        uint32
	MaxVRFs       uint32
	MaxTunnels    uint32
	MaxNeighbors  uint32
	RingBufSize   uint32
}

// DefaultConfig provides standard sizing
func DefaultConfig() Config {
	return Config{
		PinPath:       DefaultBPFFSPath,
		MaxRoutes:     65536,
		MaxNextHops:   16384,
		MaxServices:   16384,
		MaxEndpoints:  65536,
		MaxConntrack:  262144,
		MaxPolicies:   32768,
		MaxIdentities: 65536,
		MaxNAT:        131072,
		MaxVRFs:       1024,
		MaxTunnels:    4096,
		MaxNeighbors:  16384,
		RingBufSize:   1048576, // 1MB buffer
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

	// 4. Endpoints (Hash)
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

	// 8. Identities (Hash: IP -> Security Identity ID)
	idSpec := &ebpf.MapSpec{
		Name:       IdentityMapName,
		Type:       ebpf.Hash,
		KeySize:    4, // IP (uint32)
		ValueSize:  4, // Security ID (uint32)
		MaxEntries: cfg.MaxIdentities,
	}
	idMap, err := ebpf.NewMap(idSpec)
	if err != nil {
		return fmt.Errorf("create identity map: %w", err)
	}
	m.identities = idMap

	// 9. NAT (LRU Hash: FlowKey -> NATEntry)
	natSpec := &ebpf.MapSpec{
		Name:       NATMapName,
		Type:       ebpf.LRUHash,
		KeySize:    16, // struct flow_key
		ValueSize:  16, // struct nat_entry
		MaxEntries: cfg.MaxNAT,
	}
	natMap, err := ebpf.NewMap(natSpec)
	if err != nil {
		return fmt.Errorf("create nat map: %w", err)
	}
	m.nat = natMap

	// 10. VRF (Hash: VRF ID -> VRFConfig)
	vrfSpec := &ebpf.MapSpec{
		Name:       VRFMapName,
		Type:       ebpf.Hash,
		KeySize:    4,  // VRF ID (uint32)
		ValueSize:  12, // struct vrf_config
		MaxEntries: cfg.MaxVRFs,
	}
	vrfMap, err := ebpf.NewMap(vrfSpec)
	if err != nil {
		return fmt.Errorf("create vrf map: %w", err)
	}
	m.vrf = vrfMap

	// 11. Tunnels (Hash: Tunnel ID -> TunnelInfo)
	tunSpec := &ebpf.MapSpec{
		Name:       TunnelMapName,
		Type:       ebpf.Hash,
		KeySize:    4,  // Tunnel ID (uint32)
		ValueSize:  16, // struct tunnel_info
		MaxEntries: cfg.MaxTunnels,
	}
	tunMap, err := ebpf.NewMap(tunSpec)
	if err != nil {
		return fmt.Errorf("create tunnel map: %w", err)
	}
	m.tunnels = tunMap

	// 12. Neighbors (Hash: IP -> MAC)
	neighSpec := &ebpf.MapSpec{
		Name:       NeighborMapName,
		Type:       ebpf.LRUHash,
		KeySize:    4,  // IP (uint32)
		ValueSize:  12, // struct neighbor_entry
		MaxEntries: cfg.MaxNeighbors,
	}
	neighMap, err := ebpf.NewMap(neighSpec)
	if err != nil {
		return fmt.Errorf("create neighbor map: %w", err)
	}
	m.neighbors = neighMap

	// 13. Event Ring Buffer
	ringSpec := &ebpf.MapSpec{
		Name:       EventRingBufferName,
		Type:       ebpf.RingBuf,
		MaxEntries: cfg.RingBufSize,
	}
	ringMap, err := ebpf.NewMap(ringSpec)
	if err != nil {
		return fmt.Errorf("create ring buffer map: %w", err)
	}
	m.events = ringMap

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
		RouteMapName:        m.routes,
		NextHopMapName:      m.nexthops,
		ServiceMapName:      m.services,
		EndpointMapName:     m.endpoints,
		ConntrackMapName:    m.conntrack,
		PolicyMapName:       m.policies,
		CounterMapName:      m.counters,
		IdentityMapName:     m.identities,
		NATMapName:          m.nat,
		VRFMapName:          m.vrf,
		TunnelMapName:       m.tunnels,
		NeighborMapName:     m.neighbors,
		EventRingBufferName: m.events,
	}

	for name, bpfMap := range maps {
		if bpfMap == nil {
			continue
		}
		target := filepath.Join(m.pinPath, name)
		_ = os.Remove(target) // remove stale pin if present
		if err := bpfMap.Pin(target); err != nil {
			return fmt.Errorf("failed to pin %s: %w", name, err)
		}
	}
	return nil
}

// Routes
func (m *MapManager) AddRoute(key types.LPMV4Key, nextHopID uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.routes.Put(key, nextHopID)
}

func (m *MapManager) DeleteRoute(key types.LPMV4Key) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.routes.Delete(key)
}

func (m *MapManager) LookupRoute(key types.LPMV4Key) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var nextHopID uint32
	if err := m.routes.Lookup(key, &nextHopID); err != nil {
		return 0, err
	}
	return nextHopID, nil
}

// NextHops
func (m *MapManager) AddNextHop(id uint32, nh types.NextHopInfo) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nexthops.Put(id, nh)
}

func (m *MapManager) DeleteNextHop(id uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nexthops.Delete(id)
}

func (m *MapManager) LookupNextHop(id uint32) (types.NextHopInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var nh types.NextHopInfo
	if err := m.nexthops.Lookup(id, &nh); err != nil {
		return types.NextHopInfo{}, err
	}
	return nh, nil
}

// Identities
func (m *MapManager) SetIdentity(ip uint32, identityID uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.identities.Put(ip, identityID)
}

func (m *MapManager) LookupIdentity(ip uint32) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var identityID uint32
	if err := m.identities.Lookup(ip, &identityID); err != nil {
		return 0, err
	}
	return identityID, nil
}

func (m *MapManager) DeleteIdentity(ip uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.identities.Delete(ip)
}

// VRF
func (m *MapManager) SetVRF(id uint32, cfg types.VRFConfig) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.vrf.Put(id, cfg)
}

func (m *MapManager) LookupVRF(id uint32) (types.VRFConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var cfg types.VRFConfig
	if err := m.vrf.Lookup(id, &cfg); err != nil {
		return types.VRFConfig{}, err
	}
	return cfg, nil
}

func (m *MapManager) DeleteVRF(id uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.vrf.Delete(id)
}

// Tunnels
func (m *MapManager) SetTunnel(id uint32, info types.TunnelInfo) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tunnels.Put(id, info)
}

func (m *MapManager) LookupTunnel(id uint32) (types.TunnelInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var info types.TunnelInfo
	if err := m.tunnels.Lookup(id, &info); err != nil {
		return types.TunnelInfo{}, err
	}
	return info, nil
}

func (m *MapManager) DeleteTunnel(id uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tunnels.Delete(id)
}

// Neighbors
func (m *MapManager) SetNeighbor(ip uint32, entry types.NeighborEntry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.neighbors.Put(ip, entry)
}

func (m *MapManager) LookupNeighbor(ip uint32) (types.NeighborEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var entry types.NeighborEntry
	if err := m.neighbors.Lookup(ip, &entry); err != nil {
		return types.NeighborEntry{}, err
	}
	return entry, nil
}

func (m *MapManager) DeleteNeighbor(ip uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.neighbors.Delete(ip)
}

// Services (ClusterIP VIP mapping)
func (m *MapManager) AddService(key types.ServiceKey, endpointID uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services.Put(key, endpointID)
}

func (m *MapManager) LookupService(key types.ServiceKey) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var epID uint32
	if err := m.services.Lookup(key, &epID); err != nil {
		return 0, err
	}
	return epID, nil
}

func (m *MapManager) DeleteService(key types.ServiceKey) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services.Delete(key)
}

// Endpoints (Backend Pod Targets)
func (m *MapManager) AddEndpoint(id uint32, ep types.ServiceEndpoint) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.endpoints.Put(id, ep)
}

func (m *MapManager) LookupEndpoint(id uint32) (types.ServiceEndpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ep types.ServiceEndpoint
	if err := m.endpoints.Lookup(id, &ep); err != nil {
		return types.ServiceEndpoint{}, err
	}
	return ep, nil
}

func (m *MapManager) DeleteEndpoint(id uint32) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.endpoints.Delete(id)
}

// Conntrack (Stateful Flow Tracking)
func (m *MapManager) SetConntrack(key types.FlowKey, entry types.FlowEntry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conntrack.Put(key, entry)
}

func (m *MapManager) LookupConntrack(key types.FlowKey) (types.FlowEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var entry types.FlowEntry
	if err := m.conntrack.Lookup(key, &entry); err != nil {
		return types.FlowEntry{}, err
	}
	return entry, nil
}

func (m *MapManager) DeleteConntrack(key types.FlowKey) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conntrack.Delete(key)
}

// NAT (L4 Translation State)
func (m *MapManager) SetNAT(key types.FlowKey, entry types.NATEntry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nat.Put(key, entry)
}

func (m *MapManager) LookupNAT(key types.FlowKey) (types.NATEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var entry types.NATEntry
	if err := m.nat.Lookup(key, &entry); err != nil {
		return types.NATEntry{}, err
	}
	return entry, nil
}

func (m *MapManager) DeleteNAT(key types.FlowKey) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nat.Delete(key)
}

// Policies
func (m *MapManager) SetPolicy(key types.PolicyKey, action types.PolicyAction) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.policies.Put(key, uint32(action))
}

func (m *MapManager) DeletePolicy(key types.PolicyKey) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.policies.Delete(key)
}

// Counters (Per-CPU aggregation)
func (m *MapManager) GetCounter(index uint32) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.counters == nil {
		return 0, nil
	}
	var perCPUValues []uint64
	if err := m.counters.Lookup(index, &perCPUValues); err != nil {
		return 0, err
	}
	var sum uint64
	for _, v := range perCPUValues {
		sum += v
	}
	return sum, nil
}

// RingBuffer handle
func (m *MapManager) EventsMap() *ebpf.Map {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events
}

// Close releases all loaded BPF map handles
func (m *MapManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	maps := []*ebpf.Map{
		m.routes, m.nexthops, m.services, m.endpoints,
		m.conntrack, m.policies, m.counters, m.identities,
		m.nat, m.vrf, m.tunnels, m.neighbors, m.events,
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
