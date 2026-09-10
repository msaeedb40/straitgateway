// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package wireguard manages WireGuard tunnel interfaces for
// transparent pod-to-pod and transit encryption.
//
// Uses Linux kernel WireGuard via netlink (no userspace daemon).
// Key generation uses Curve25519.
package wireguard

import (
	"crypto/rand"
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/curve25519"
)

const (
	// InterfaceName is the WireGuard interface for pod-to-pod encryption.
	InterfaceName = "sg-wg0"
	// TransitInterfaceName is the WireGuard interface for transit tunnels.
	TransitInterfaceName = "sg-wg-transit"
	// ListenPort is the default WireGuard listen port.
	ListenPort = 51871
)

// KeyPair holds a WireGuard Curve25519 key pair.
type KeyPair struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
}

// Manager manages WireGuard interfaces and peer configurations.
type Manager struct {
	log   *zap.Logger
	mu    sync.Mutex
	keys  *KeyPair
	peers map[string]*Peer // keyed by tunnel IP
}

// Peer represents a WireGuard peer (remote node or cluster).
type Peer struct {
	PublicKey    [32]byte
	Endpoint     netip.AddrPort
	AllowedCIDRs []netip.Prefix
	// Persistent keepalive interval in seconds (0 = disabled).
	KeepaliveInterval int
}

// NewManager creates a new WireGuard manager.
func NewManager(log *zap.Logger) (*Manager, error) {
	keys, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generating WireGuard keys: %w", err)
	}
	log.Info("WireGuard key pair generated",
		zap.String("publicKey", fmt.Sprintf("%x", keys.PublicKey[:8])),
	)
	return &Manager{
		log:   log,
		keys:  keys,
		peers: make(map[string]*Peer),
	}, nil
}

// GenerateKeyPair generates a new Curve25519 key pair.
func GenerateKeyPair() (*KeyPair, error) {
	var privKey [32]byte
	if _, err := rand.Read(privKey[:]); err != nil {
		return nil, err
	}
	// Clamp private key per Curve25519 spec.
	privKey[0] &= 248
	privKey[31] &= 127
	privKey[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}, nil
}

// PublicKey returns the node's WireGuard public key.
func (m *Manager) PublicKey() [32]byte {
	return m.keys.PublicKey
}

// AddPeer adds or updates a WireGuard peer.
func (m *Manager) AddPeer(tunnelIP string, peer *Peer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.peers[tunnelIP] = peer
	m.log.Info("WireGuard peer added",
		zap.String("tunnelIP", tunnelIP),
		zap.String("endpoint", peer.Endpoint.String()),
		zap.Int("allowedCIDRs", len(peer.AllowedCIDRs)),
	)

	// TODO: Apply via netlink WireGuard API
	return nil
}

// RemovePeer removes a WireGuard peer.
func (m *Manager) RemovePeer(tunnelIP string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.peers, tunnelIP)
	m.log.Info("WireGuard peer removed", zap.String("tunnelIP", tunnelIP))

	// TODO: Remove via netlink WireGuard API
	return nil
}
