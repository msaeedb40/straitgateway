// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package bgp implements BGP-4 peering for route advertisement and learning.
// Used for advertising LoadBalancer IPs, pod CIDRs, and service VIPs
// to upstream routers, and for learning external routes.
//
// BFD integration provides sub-second link failure detection.
package bgp

import (
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"
)

// SessionState represents the BGP Finite State Machine state.
type SessionState string

const (
	StateIdle        SessionState = "Idle"
	StateConnect     SessionState = "Connect"
	StateActive      SessionState = "Active"
	StateOpenSent    SessionState = "OpenSent"
	StateOpenConfirm SessionState = "OpenConfirm"
	StateEstablished SessionState = "Established"
)

// Session represents a BGP peering session with a single peer.
type Session struct {
	mu     sync.RWMutex
	log    *zap.Logger

	PeerAddress  netip.Addr
	PeerASN      uint32
	LocalASN     uint32
	LocalAddress netip.Addr
	State        SessionState

	// PrefixesAdvertised tracks prefixes we advertise to this peer.
	PrefixesAdvertised []netip.Prefix
	// PrefixesReceived tracks prefixes received from this peer.
	PrefixesReceived   []netip.Prefix

	// BFD integration for fast failover.
	BFDEnabled bool
	BFDState   string // "Up", "Down", "Init", "AdminDown"
}

// Manager manages all BGP peering sessions.
type Manager struct {
	log      *zap.Logger
	mu       sync.RWMutex
	sessions map[string]*Session // keyed by peer address
	localASN uint32
}

// NewManager creates a new BGP manager.
func NewManager(localASN uint32, log *zap.Logger) *Manager {
	return &Manager{
		log:      log,
		sessions: make(map[string]*Session),
		localASN: localASN,
	}
}

// AddPeer creates a new BGP peering session.
func (m *Manager) AddPeer(peerAddr netip.Addr, peerASN uint32) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := peerAddr.String()
	if _, exists := m.sessions[key]; exists {
		return nil, fmt.Errorf("BGP peer %s already exists", key)
	}

	sess := &Session{
		log:         m.log.With(zap.String("peer", key)),
		PeerAddress: peerAddr,
		PeerASN:     peerASN,
		LocalASN:    m.localASN,
		State:       StateIdle,
	}

	m.sessions[key] = sess
	m.log.Info("BGP peer added",
		zap.String("peer", key),
		zap.Uint32("peerASN", peerASN),
		zap.Uint32("localASN", m.localASN),
	)

	return sess, nil
}

// RemovePeer removes a BGP peering session.
func (m *Manager) RemovePeer(peerAddr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[peerAddr]; !exists {
		return fmt.Errorf("BGP peer %s not found", peerAddr)
	}

	delete(m.sessions, peerAddr)
	m.log.Info("BGP peer removed", zap.String("peer", peerAddr))
	return nil
}

// AdvertisePrefix adds a prefix to advertise to a specific peer or all peers.
func (m *Manager) AdvertisePrefix(prefix netip.Prefix) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sess := range m.sessions {
		sess.mu.Lock()
		sess.PrefixesAdvertised = append(sess.PrefixesAdvertised, prefix)
		sess.mu.Unlock()
	}
	m.log.Info("prefix advertised to all peers", zap.String("prefix", prefix.String()))
}

// AllSessions returns all current BGP sessions.
func (m *Manager) AllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}
