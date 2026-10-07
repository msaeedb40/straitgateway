package bfd

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// SessionState defines BFD state
type SessionState string

const (
	StateAdminDown SessionState = "AdminDown"
	StateDown      SessionState = "Down"
	StateInit      SessionState = "Init"
	StateUp        SessionState = "Up"
)

// SessionConfig defines BFD timing parameters
type SessionConfig struct {
	PeerIP           net.IP
	LocalIP          net.IP
	DesiredMinTxMs   uint32
	RequiredMinRxMs  uint32
	DetectMultiplier uint8
}

// Session represents an active BFD session
type Session struct {
	Config      SessionConfig
	State       SessionState
	LastRx      time.Time
	OnDownState func(peerIP net.IP)
}

// Monitor coordinates BFD sessions for fast failure detection
type Monitor struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	stopCh   chan struct{}
}

// NewMonitor initializes a BFD Monitor
func NewMonitor() *Monitor {
	return &Monitor{
		sessions: make(map[string]*Session),
		stopCh:   make(chan struct{}),
	}
}

// AddSession registers and activates BFD monitoring for a peer
func (m *Monitor) AddSession(cfg SessionConfig, onDown func(peerIP net.IP)) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cfg.DesiredMinTxMs == 0 {
		cfg.DesiredMinTxMs = 250 // 250ms default
	}
	if cfg.RequiredMinRxMs == 0 {
		cfg.RequiredMinRxMs = 250
	}
	if cfg.DetectMultiplier == 0 {
		cfg.DetectMultiplier = 3 // 750ms total detection time
	}

	session := &Session{
		Config:      cfg,
		State:       StateUp,
		LastRx:      time.Now(),
		OnDownState: onDown,
	}

	m.sessions[cfg.PeerIP.String()] = session
	return session
}

// TriggerFailure simulates or handles a missed heartbeat / link failure
func (m *Monitor) TriggerFailure(peerIP net.IP) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[peerIP.String()]
	if !exists {
		return fmt.Errorf("session for %s not found", peerIP.String())
	}

	s.State = StateDown
	if s.OnDownState != nil {
		// Immediately trigger route invalidation callback
		s.OnDownState(peerIP)
	}
	return nil
}

// ReceiveHeartbeat records an incoming BFD control packet / echo
func (m *Monitor) ReceiveHeartbeat(peerIP net.IP) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[peerIP.String()]
	if !exists {
		return false
	}

	s.LastRx = time.Now()
	if s.State != StateUp && s.State != StateAdminDown {
		s.State = StateUp
	}
	return true
}

// CheckTimeouts inspects all sessions and marks timed-out peers as Down
func (m *Monitor) CheckTimeouts() []net.IP {
	m.mu.Lock()
	defer m.mu.Unlock()

	var failedPeers []net.IP
	now := time.Now()

	for _, s := range m.sessions {
		if s.State != StateUp {
			continue
		}
		timeout := time.Duration(s.Config.RequiredMinRxMs*uint32(s.Config.DetectMultiplier)) * time.Millisecond
		if now.Sub(s.LastRx) > timeout {
			s.State = StateDown
			failedPeers = append(failedPeers, s.Config.PeerIP)
			if s.OnDownState != nil {
				s.OnDownState(s.Config.PeerIP)
			}
		}
	}
	return failedPeers
}

// RemoveSession deletes a BFD session
func (m *Monitor) RemoveSession(peerIP net.IP) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, peerIP.String())
}

// ListSessions returns all active BFD sessions
func (m *Monitor) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		list = append(list, s)
	}
	return list
}

// GetSession retrieves the current state of a peer's BFD session
func (m *Monitor) GetSession(peerIP net.IP) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, exists := m.sessions[peerIP.String()]
	return s, exists
}

// Close terminates monitor
func (m *Monitor) Close() {
	close(m.stopCh)
}

