package routing

import (
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// Config configures the Routing Dataplane Manager
type Config struct {
	LocalASN      uint32 `json:"localAsn"`
	RouterID      net.IP `json:"routerId"`
	WireGuardDev  string `json:"wireguardDev"`
	WireGuardPort int    `json:"wireguardPort"`
}

// DefaultConfig provides standard routing dataplane defaults
func DefaultConfig() Config {
	return Config{
		LocalASN:      65001,
		RouterID:      net.ParseIP("10.0.0.1"),
		WireGuardDev:  "wg0",
		WireGuardPort: 51820,
	}
}

// Manager coordinates FIB, BGP, BFD, and WireGuard with the eBPF LPM routing dataplane
type Manager struct {
	mu           sync.RWMutex
	cfg          Config
	mapMgr       *maps.MapManager
	fib          *routing.FIBEngine
	routeMgr     *routing.RouteManager
	bgpSpeaker   *bgp.Speaker
	bfdMonitor   *bfd.Monitor
	wireguardMgr *wireguard.Manager
}

// NewManager creates a new Routing Dataplane Manager
func NewManager(cfg Config, mapMgr *maps.MapManager) (*Manager, error) {
	if cfg.LocalASN == 0 {
		cfg.LocalASN = 65001
	}
	if cfg.RouterID == nil {
		cfg.RouterID = net.ParseIP("10.0.0.1")
	}
	if cfg.WireGuardDev == "" {
		cfg.WireGuardDev = "wg0"
	}
	if cfg.WireGuardPort == 0 {
		cfg.WireGuardPort = 51820
	}

	fib := routing.NewFIBEngine()
	routeMgr := routing.NewRouteManager(fib, mapMgr)

	bgpSpeaker := bgp.NewSpeaker(cfg.LocalASN, cfg.RouterID, routeMgr)
	bfdMonitor := bfd.NewMonitor()
	bgpSpeaker.SetBFDMonitor(bfdMonitor)

	wgMgr, _ := wireguard.NewManager(cfg.WireGuardDev, cfg.WireGuardPort)
	if wgMgr != nil {
		wgMgr.SetRouteManager(routeMgr)
	}

	return &Manager{
		cfg:          cfg,
		mapMgr:       mapMgr,
		fib:          fib,
		routeMgr:     routeMgr,
		bgpSpeaker:   bgpSpeaker,
		bfdMonitor:   bfdMonitor,
		wireguardMgr: wgMgr,
	}, nil
}

// FIB returns the underlying FIB Engine
func (m *Manager) FIB() *routing.FIBEngine {
	return m.fib
}

// RouteManager returns the RouteManager
func (m *Manager) RouteManager() *routing.RouteManager {
	return m.routeMgr
}

// BGPSpeaker returns the BGP Speaker
func (m *Manager) BGPSpeaker() *bgp.Speaker {
	return m.bgpSpeaker
}

// BFDMonitor returns the BFD Monitor
func (m *Manager) BFDMonitor() *bfd.Monitor {
	return m.bfdMonitor
}

// WireGuardManager returns the WireGuard Manager
func (m *Manager) WireGuardManager() *wireguard.Manager {
	return m.wireguardMgr
}

// AddVRF registers a new tenant VRF routing domain
func (m *Manager) AddVRF(id uint32, name, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fib.AddVRF(id, name, description)

	if m.mapMgr != nil {
		vrfCfg := types.VRFConfig{
			ID:      id,
			TableID: id + 1000,
			Flags:   1,
		}
		return m.mapMgr.SetVRF(id, vrfCfg)
	}
	return nil
}

// InstallRoute synchronizes a route entry into the FIB and eBPF LPM trie
func (m *Manager) InstallRoute(dest net.IPNet, nextHopIP net.IP, vrfID uint32, ifName string, precedence routing.RoutePrecedence) error {
	entry := &routing.RouteEntry{
		Destination: dest,
		VRFID:       vrfID,
		NextHopIP:   nextHopIP,
		Interface:   ifName,
		Precedence:  precedence,
	}
	return m.routeMgr.SyncRoute(entry)
}

// WithdrawRoute removes a route entry from the FIB and eBPF LPM trie
func (m *Manager) WithdrawRoute(vrfID uint32, dest net.IPNet) error {
	return m.routeMgr.WithdrawRoute(vrfID, dest)
}

// SetupBFDPeering configures sub-50ms BFD failure detection for a BGP peer
func (m *Manager) SetupBFDPeering(peerIP net.IP, txIntervalMs, rxIntervalMs uint32) *bfd.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := bfd.SessionConfig{
		PeerIP:           peerIP,
		LocalIP:          m.cfg.RouterID,
		DesiredMinTxMs:   txIntervalMs,
		RequiredMinRxMs:  rxIntervalMs,
		DetectMultiplier: 3,
	}

	// On BFD down, immediately retract routes to that next hop in the eBPF dataplane
	onDown := func(downPeer net.IP) {
		_, _ = m.routeMgr.WithdrawNextHop(downPeer)
	}

	return m.bfdMonitor.AddSession(cfg, onDown)
}

// Close gracefully closes routing services
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.wireguardMgr != nil {
		_ = m.wireguardMgr.Close()
	}
	return nil
}
