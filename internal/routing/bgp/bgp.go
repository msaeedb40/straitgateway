package bgp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/routing"
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

// AddNeighbor registers a new BGP neighbor
func (s *Speaker) AddNeighbor(cfg PeerConfig) (*Peer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%d", cfg.NeighborIP.String(), cfg.VRFID)
	peer := &Peer{
		Config:          cfg,
		State:           StateEstablished, // Mark established on connect
		EstablishedTime: time.Now(),
		ReceivedRoutes:  make([]net.IPNet, 0),
		AdvertisedRoutes: cfg.AdvertisedPrefix,
	}

	s.peers[key] = peer
	return peer, nil
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

// GetPeer retrieves status for a peer
func (s *Speaker) GetPeer(neighborIP net.IP, vrfID uint32) (*Peer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", neighborIP.String(), vrfID)
	peer, exists := s.peers[key]
	return peer, exists
}
