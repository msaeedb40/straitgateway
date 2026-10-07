package bgp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
)

// SessionState defines BGP state machine status
type SessionState string

const (
	StateIdle        SessionState = "Idle"
	StateConnect     SessionState = "Connect"
	StateActive      SessionState = "Active"
	StateOpenSent    SessionState = "OpenSent"
	StateOpenConfirm SessionState = "OpenConfirm"
	StateEstablished SessionState = "Established"
)

// PeerConfig holds BGP neighbor configuration
type PeerConfig struct {
	NeighborIP       net.IP
	RemoteASN        uint32
	LocalASN         uint32
	RouterID         net.IP
	HoldTimeSeconds  uint32
	EnableBFD        bool
	AuthPassword     string
	AdvertisedPrefix []net.IPNet
	VRFID            uint32
}

// Peer represents an active BGP neighbor session
type Peer struct {
	Config          PeerConfig
	State           SessionState
	EstablishedTime time.Time
	ReceivedRoutes  []net.IPNet
	AdvertisedRoutes []net.IPNet
}

// Speaker manages dynamic BGP peering sessions
type Speaker struct {
	mu           sync.RWMutex
	localASN     uint32
	routerID     net.IP
	peers        map[string]*Peer // key: neighborIP:vrf
	routeManager *routing.RouteManager
	bfdMonitor   *bfd.Monitor
}

// NewSpeaker creates a BGP speaker
func NewSpeaker(localASN uint32, routerID net.IP, rm *routing.RouteManager) *Speaker {
	return &Speaker{
		localASN:     localASN,
		routerID:     routerID,
		peers:        make(map[string]*Peer),
		routeManager: rm,
	}
}

// SetBFDMonitor links a BFD failure detector to the BGP speaker
func (s *Speaker) SetBFDMonitor(bm *bfd.Monitor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bfdMonitor = bm
}


// AddNeighbor registers a new BGP neighbor and attaches BFD if enabled
func (s *Speaker) AddNeighbor(cfg PeerConfig) (*Peer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", cfg.NeighborIP.String(), cfg.VRFID)
	peer := &Peer{
		Config:           cfg,
		State:            StateEstablished, // Mark established on connect
		EstablishedTime:  time.Now(),
		ReceivedRoutes:   make([]net.IPNet, 0),
		AdvertisedRoutes: cfg.AdvertisedPrefix,
	}

	s.peers[key] = peer

	// If BFD is requested and monitor is configured, attach session
	if cfg.EnableBFD && s.bfdMonitor != nil {
		bfdCfg := bfd.SessionConfig{
			PeerIP:  cfg.NeighborIP,
			LocalIP: s.routerID,
		}
		s.bfdMonitor.AddSession(bfdCfg, func(peerIP net.IP) {
			s.handlePeerFailure(peerIP, cfg.VRFID)
		})
	}

	return peer, nil
}

// handlePeerFailure is triggered immediately upon BFD down event
func (s *Speaker) handlePeerFailure(neighborIP net.IP, vrfID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	if !exists {
		return
	}

	peer.State = StateActive // Mark session degraded / recovering

	// Withdraw routes learned from this neighbor immediately
	if s.routeManager != nil {
		_, _ = s.routeManager.WithdrawNextHop(neighborIP)
	}
}

// RemoveNeighbor removes a BGP neighbor and withdraws learned routes
func (s *Speaker) RemoveNeighbor(neighborIP net.IP, vrfID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	if !exists {
		return fmt.Errorf("neighbor %s not found", neighborIP.String())
	}

	// Withdraw learned routes if RouteManager is attached
	if s.routeManager != nil {
		for _, r := range peer.ReceivedRoutes {
			_ = s.routeManager.WithdrawRoute(vrfID, r)
		}
	}

	// Remove from BFD monitor if attached
	if peer.Config.EnableBFD && s.bfdMonitor != nil {
		s.bfdMonitor.RemoveSession(neighborIP)
	}

	delete(s.peers, key)
	return nil
}

// ReceiveRoute injects a dynamically learned route from a peer into the FIB
func (s *Speaker) ReceiveRoute(neighborIP net.IP, vrfID uint32, prefix net.IPNet, metric uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	if !exists {
		return fmt.Errorf("neighbor %s not found", neighborIP.String())
	}

	peer.ReceivedRoutes = append(peer.ReceivedRoutes, prefix)

	if s.routeManager != nil {
		entry := &routing.RouteEntry{
			Destination: prefix,
			VRFID:       vrfID,
			NextHopIP:   neighborIP,
			Precedence:  routing.PrecedenceBGPRoute,
			Metric:      metric,
		}
		return s.routeManager.SyncRoute(entry)
	}
	return nil
}

// WithdrawRoute removes a specific route received from a neighbor
func (s *Speaker) WithdrawRoute(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	if !exists {
		return fmt.Errorf("neighbor %s not found", neighborIP.String())
	}

	// Remove from ReceivedRoutes
	prefixStr := prefix.String()
	newRoutes := make([]net.IPNet, 0, len(peer.ReceivedRoutes))
	for _, r := range peer.ReceivedRoutes {
		if r.String() != prefixStr {
			newRoutes = append(newRoutes, r)
		}
	}
	peer.ReceivedRoutes = newRoutes

	if s.routeManager != nil {
		return s.routeManager.WithdrawRoute(vrfID, prefix)
	}
	return nil
}

// AdvertisePrefix records a local prefix advertised to a neighbor
func (s *Speaker) AdvertisePrefix(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	if !exists {
		return fmt.Errorf("neighbor %s not found", neighborIP.String())
	}

	peer.AdvertisedRoutes = append(peer.AdvertisedRoutes, prefix)
	return nil
}

// ListPeers returns all configured BGP peers
func (s *Speaker) ListPeers() []*Peer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Peer, 0, len(s.peers))
	for _, p := range s.peers {
		list = append(list, p)
	}
	return list
}

// GetPeer retrieves status for a peer
func (s *Speaker) GetPeer(neighborIP net.IP, vrfID uint32) (*Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	return peer, exists
}

