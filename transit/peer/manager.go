// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package peer manages WireGuard peer relationships between transit gateway nodes.
// The peer manager drives the encryption/wireguard package to establish tunnels.
package peer

import (
	"fmt"
	"net/netip"
	"sync"
	"time"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"go.uber.org/zap"
)

// PeerState represents the state of a WireGuard peer connection.
type PeerState string

const (
	PeerStatePending    PeerState = "pending"
	PeerStateConnected  PeerState = "connected"
	PeerStateDisconnected PeerState = "disconnected"
)

// Peer represents a remote transit peer.
type Peer struct {
	ClusterID   sgtypes.ClusterID
	GatewayName string
	Endpoint    netip.AddrPort
	WGPublicKey [32]byte
	AllowedCIDRs []netip.Prefix
	State       PeerState
	LastSeen    time.Time
}

// Manager manages transit WireGuard peer connections.
type Manager struct {
	log   *zap.Logger
	mu    sync.RWMutex
	peers map[string]*Peer // keyed by ClusterID:GatewayName
}

// New creates a new peer Manager.
func New(log *zap.Logger) *Manager {
	return &Manager{log: log, peers: make(map[string]*Peer)}
}

func peerKey(clusterID sgtypes.ClusterID, gatewayName string) string {
	return fmt.Sprintf("%d:%s", clusterID, gatewayName)
}

// AddPeer adds or updates a transit peer.
func (m *Manager) AddPeer(p *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := peerKey(p.ClusterID, p.GatewayName)
	p.State = PeerStatePending
	m.peers[key] = p
	m.log.Info("transit peer added",
		zap.String("gateway", p.GatewayName),
		zap.String("endpoint", p.Endpoint.String()),
		zap.Uint32("cluster", uint32(p.ClusterID)),
	)
}

// RemovePeer removes a transit peer.
func (m *Manager) RemovePeer(clusterID sgtypes.ClusterID, gatewayName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, peerKey(clusterID, gatewayName))
}

// SetState updates a peer's connection state.
func (m *Manager) SetState(clusterID sgtypes.ClusterID, gatewayName string, state PeerState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := peerKey(clusterID, gatewayName)
	if p, ok := m.peers[key]; ok {
		p.State = state
		p.LastSeen = time.Now()
	}
}

// All returns all peers.
func (m *Manager) All() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Peer, 0, len(m.peers))
	for _, p := range m.peers {
		out = append(out, p)
	}
	return out
}

// Connected returns all peers in the connected state.
func (m *Manager) Connected() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Peer
	for _, p := range m.peers {
		if p.State == PeerStateConnected {
			out = append(out, p)
		}
	}
	return out
}
