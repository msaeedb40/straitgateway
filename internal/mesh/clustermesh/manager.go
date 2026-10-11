package clustermesh

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

// Manager coordinates multi-cluster peering, cross-cluster transit routing,
// WireGuard mesh links, and global service discovery.
type Manager struct {
	mu             sync.RWMutex
	cfg            Config
	wireguardMgr   *wireguard.Manager
	routeMgr       *routing.RouteManager
	peers          map[string]*ClusterPeerState      // clusterID -> state
	globalServices map[string]*GlobalServiceRecord   // globalName -> record
	nextVIPSeq     uint32
	globalCIDRNet  net.IPNet
}

// NewManager creates a new ClusterMesh manager
func NewManager(cfg Config, wgMgr *wireguard.Manager, routeMgr *routing.RouteManager) (*Manager, error) {
	if cfg.LocalClusterID == "" {
		cfg.LocalClusterID = "cluster-local"
	}
	if cfg.GlobalCIDR == "" {
		cfg.GlobalCIDR = "240.0.0.0/8"
	}
	if cfg.DomainSuffix == "" {
		cfg.DomainSuffix = "global.straitgateway"
	}
	if cfg.DNSPort == 0 {
		cfg.DNSPort = 5353
	}

	_, cidrNet, err := net.ParseCIDR(cfg.GlobalCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid global CIDR %s: %w", cfg.GlobalCIDR, err)
	}

	return &Manager{
		cfg:            cfg,
		wireguardMgr:   wgMgr,
		routeMgr:       routeMgr,
		peers:          make(map[string]*ClusterPeerState),
		globalServices: make(map[string]*GlobalServiceRecord),
		nextVIPSeq:     1,
		globalCIDRNet:  *cidrNet,
	}, nil
}

// AddClusterPeer registers a remote Kubernetes cluster, programs the WireGuard peer,
// and installs transit routes for remote Pod and Service CIDRs.
func (m *Manager) AddClusterPeer(peer *v1alpha1.ClusterPeering) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer == nil || peer.Spec.ClusterID == "" {
		return fmt.Errorf("cluster peering spec must contain a non-empty ClusterID")
	}

	var podNets []net.IPNet
	for _, cidr := range peer.Spec.PodCIDRs {
		_, parsed, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid pod CIDR %q: %w", cidr, err)
		}
		podNets = append(podNets, *parsed)
	}

	var svcNets []net.IPNet
	for _, cidr := range peer.Spec.ServiceCIDRs {
		_, parsed, err := net.ParseCIDR(cidr)
		if err == nil {
			svcNets = append(svcNets, *parsed)
		}
	}

	// 1. Program WireGuard peer if WireGuard Manager is attached
	if m.wireguardMgr != nil && peer.Spec.WireGuardPublicKey != "" {
		var allowedIPs []net.IPNet
		allowedIPs = append(allowedIPs, podNets...)
		allowedIPs = append(allowedIPs, svcNets...)

		wgPeer := wireguard.PeerConfig{
			PublicKey:           peer.Spec.WireGuardPublicKey,
			Endpoint:            peer.Spec.WireGuardEndpoint,
			AllowedIPs:         allowedIPs,
			PersistentKeepalive: int(peer.Spec.PersistentKeepaliveSeconds),
		}
		if wgPeer.PersistentKeepalive == 0 {
			wgPeer.PersistentKeepalive = 25
		}

		if err := m.wireguardMgr.AddPeer(wgPeer); err != nil {
			return fmt.Errorf("failed to register WireGuard peer: %w", err)
		}
	}

	// 2. Install FIB routes into RouteManager
	routeCount := 0
	if m.routeMgr != nil {
		for _, pNet := range podNets {
			_ = m.routeMgr.SyncRoute(&routing.RouteEntry{
				Destination: pNet,
				VRFID:       peer.Spec.VRFID,
				Interface:   "wg0",
				Precedence:  routing.PrecedenceClusterRoute,
			})
			routeCount++
		}
		for _, sNet := range svcNets {
			_ = m.routeMgr.SyncRoute(&routing.RouteEntry{
				Destination: sNet,
				VRFID:       peer.Spec.VRFID,
				Interface:   "wg0",
				Precedence:  routing.PrecedenceClusterRoute,
			})
			routeCount++
		}
	}

	state := &ClusterPeerState{
		ClusterID:         peer.Spec.ClusterID,
		WireGuardEndpoint: peer.Spec.WireGuardEndpoint,
		WireGuardPubkey:   peer.Spec.WireGuardPublicKey,
		PodCIDRs:          podNets,
		ServiceCIDRs:      svcNets,
		VRFID:             peer.Spec.VRFID,
		Phase:             "Connected",
		LatencyMillis:     1.5,
		LastHeartbeat:     time.Now(),
		ActiveRoutes:      routeCount,
	}

	m.peers[peer.Spec.ClusterID] = state
	return nil
}

// RemoveClusterPeer tears down the cluster peering, removes WireGuard peer,
// withdraws FIB transit routes, and purges remote endpoints.
func (m *Manager) RemoveClusterPeer(clusterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.peers[clusterID]
	if !exists {
		return nil
	}

	// 1. Remove WireGuard peer
	if m.wireguardMgr != nil && state.WireGuardPubkey != "" {
		_ = m.wireguardMgr.RemovePeer(state.WireGuardPubkey)
	}

	// 2. Withdraw FIB routes
	if m.routeMgr != nil {
		for _, pNet := range state.PodCIDRs {
			_ = m.routeMgr.WithdrawRoute(state.VRFID, pNet)
		}
		for _, sNet := range state.ServiceCIDRs {
			_ = m.routeMgr.WithdrawRoute(state.VRFID, sNet)
		}
	}

	// 3. Purge cluster from global services backends
	for _, rec := range m.globalServices {
		rec.mu.Lock()
		delete(rec.ClusterBackends, clusterID)
		rec.mu.Unlock()
	}

	delete(m.peers, clusterID)
	return nil
}

// GetClusterPeer returns state for a specific peered cluster
func (m *Manager) GetClusterPeer(clusterID string) (*ClusterPeerState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.peers[clusterID]
	return st, ok
}

// ListClusterPeers returns all registered peer states
func (m *Manager) ListClusterPeers() []*ClusterPeerState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*ClusterPeerState, 0, len(m.peers))
	for _, st := range m.peers {
		res = append(res, st)
	}
	return res
}

// FormatGlobalName constructs the canonical FQDN: <svc>.<ns>.<domainSuffix>
func (m *Manager) FormatGlobalName(serviceName, namespace string) string {
	return fmt.Sprintf("%s.%s.%s", strings.ToLower(serviceName), strings.ToLower(namespace), m.cfg.DomainSuffix)
}

// allocateVIP allocates a deterministic or sequential VIP within the global CIDR
func (m *Manager) allocateVIP() net.IP {
	seq := atomic.AddUint32(&m.nextVIPSeq, 1)
	base := make(net.IP, len(m.globalCIDRNet.IP))
	copy(base, m.globalCIDRNet.IP)

	if len(base) == 16 && base.To4() != nil {
		base = base.To4()
	}
	if len(base) == 4 {
		val := binary.BigEndian.Uint32(base) + seq
		out := make(net.IP, 4)
		binary.BigEndian.PutUint32(out, val)
		return out
	}
	return net.ParseIP("240.0.0.1")
}

// ExportService advertises a local service into the ClusterMesh catalog
func (m *Manager) ExportService(serviceName, namespace string, ports []v1alpha1.GlobalServicePort, trafficPolicy string) (*GlobalServiceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if serviceName == "" || namespace == "" {
		return nil, fmt.Errorf("serviceName and namespace must be provided")
	}

	globalName := m.FormatGlobalName(serviceName, namespace)
	rec, exists := m.globalServices[globalName]
	if !exists {
		vip := m.allocateVIP()
		rec = &GlobalServiceRecord{
			ServiceName:     serviceName,
			Namespace:       namespace,
			GlobalName:      globalName,
			GlobalVIP:       vip,
			Ports:           ports,
			TrafficPolicy:   trafficPolicy,
			ClusterBackends: make(map[string][]string),
		}
		if rec.TrafficPolicy == "" {
			rec.TrafficPolicy = "ClusterMeshPreferred"
		}
		m.globalServices[globalName] = rec
	} else {
		rec.mu.Lock()
		rec.Ports = ports
		if trafficPolicy != "" {
			rec.TrafficPolicy = trafficPolicy
		}
		rec.mu.Unlock()
	}

	return rec, nil
}

// ImportService imports a remote global service definition
func (m *Manager) ImportService(req *v1alpha1.GlobalServiceImport) (*GlobalServiceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req == nil || req.Spec.GlobalServiceName == "" {
		return nil, fmt.Errorf("global service import spec cannot be empty")
	}

	globalName := strings.ToLower(req.Spec.GlobalServiceName)
	parts := strings.Split(globalName, ".")
	svcName := parts[0]
	ns := "default"
	if len(parts) > 1 {
		ns = parts[1]
	}

	var vip net.IP
	if req.Spec.GlobalVIP != "" {
		vip = net.ParseIP(req.Spec.GlobalVIP)
	}
	if vip == nil {
		vip = m.allocateVIP()
	}

	rec, exists := m.globalServices[globalName]
	if !exists {
		rec = &GlobalServiceRecord{
			ServiceName:     svcName,
			Namespace:       ns,
			GlobalName:      globalName,
			GlobalVIP:       vip,
			Ports:           req.Spec.Ports,
			TrafficPolicy:   "ClusterMeshPreferred",
			ClusterBackends: make(map[string][]string),
		}
		m.globalServices[globalName] = rec
	}

	rec.mu.Lock()
	for _, cb := range req.Spec.Clusters {
		rec.ClusterBackends[cb.ClusterID] = cb.EndpointIPs
	}
	rec.mu.Unlock()

	return rec, nil
}

// RegisterRemoteEndpoints associates backend endpoints with a service for a specific cluster
func (m *Manager) RegisterRemoteEndpoints(serviceName, namespace, clusterID string, endpoints []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	globalName := m.FormatGlobalName(serviceName, namespace)
	rec, exists := m.globalServices[globalName]
	if !exists {
		vip := m.allocateVIP()
		rec = &GlobalServiceRecord{
			ServiceName:     serviceName,
			Namespace:       namespace,
			GlobalName:      globalName,
			GlobalVIP:       vip,
			TrafficPolicy:   "ClusterMeshPreferred",
			ClusterBackends: make(map[string][]string),
		}
		m.globalServices[globalName] = rec
	}

	rec.mu.Lock()
	rec.ClusterBackends[clusterID] = endpoints
	rec.mu.Unlock()

	return nil
}

// ResolveGlobalVIP retrieves the Global VIP for an FQDN
func (m *Manager) ResolveGlobalVIP(globalName string) (net.IP, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, exists := m.globalServices[strings.ToLower(globalName)]
	if !exists {
		return nil, false
	}
	return rec.GlobalVIP, true
}

// GetGlobalService returns the service record by FQDN
func (m *Manager) GetGlobalService(globalName string) (*GlobalServiceRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, ok := m.globalServices[strings.ToLower(globalName)]
	return rec, ok
}

// ListGlobalServices returns all known global service records
func (m *Manager) ListGlobalServices() []*GlobalServiceRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]*GlobalServiceRecord, 0, len(m.globalServices))
	for _, rec := range m.globalServices {
		res = append(res, rec)
	}
	return res
}

// SelectBackend chooses an endpoint IP for a service based on configured traffic policy:
// - ClusterLocal: only endpoints belonging to clientClusterID
// - ClusterMeshPreferred: local endpoints first, fallback to remote if none available
// - RoundRobin: distributed across all available endpoints in the mesh
func (m *Manager) SelectBackend(serviceName, namespace string, clientClusterID string) (string, error) {
	m.mu.RLock()
	globalName := m.FormatGlobalName(serviceName, namespace)
	rec, exists := m.globalServices[globalName]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("global service %s not found", globalName)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()

	if clientClusterID == "" {
		clientClusterID = m.cfg.LocalClusterID
	}

	switch rec.TrafficPolicy {
	case "ClusterLocal":
		locals := rec.ClusterBackends[clientClusterID]
		if len(locals) == 0 {
			return "", fmt.Errorf("no local endpoints available in cluster %s for %s", clientClusterID, globalName)
		}
		idx := rec.rrCounter % uint64(len(locals))
		rec.rrCounter++
		return locals[idx], nil

	case "RoundRobin":
		var all []string
		for _, eps := range rec.ClusterBackends {
			all = append(all, eps...)
		}
		if len(all) == 0 {
			return "", fmt.Errorf("no backends available across mesh for %s", globalName)
		}
		idx := rec.rrCounter % uint64(len(all))
		rec.rrCounter++
		return all[idx], nil

	case "ClusterMeshPreferred":
		fallthrough
	default:
		// Try local first
		locals := rec.ClusterBackends[clientClusterID]
		if len(locals) > 0 {
			idx := rec.rrCounter % uint64(len(locals))
			rec.rrCounter++
			return locals[idx], nil
		}
		// Fallback to any remote cluster
		var remotes []string
		for cid, eps := range rec.ClusterBackends {
			if cid != clientClusterID {
				remotes = append(remotes, eps...)
			}
		}
		if len(remotes) == 0 {
			return "", fmt.Errorf("no backends available for %s", globalName)
		}
		idx := rec.rrCounter % uint64(len(remotes))
		rec.rrCounter++
		return remotes[idx], nil
	}
}
