// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package bfd implements Bidirectional Forwarding Detection (RFC 5880)
// for sub-second link failure detection in BGP peering sessions.
package bfd

import (
	"net/netip"
	"sync"
	"time"

	"go.uber.org/zap"
)

// State represents the BFD session state.
type State string

const (
	StateAdminDown State = "AdminDown"
	StateDown      State = "Down"
	StateInit      State = "Init"
	StateUp        State = "Up"
)

// Session represents a BFD session with a single peer.
type Session struct {
	mu          sync.RWMutex
	PeerAddress netip.Addr
	State       State
	// DetectMultiplier × TxInterval = detection timeout.
	DetectMultiplier uint8
	// TxInterval is the desired min transmit interval.
	TxInterval time.Duration
	// RxInterval is the required min receive interval.
	RxInterval time.Duration
	// LastPacketRx is the timestamp of the last received BFD packet.
	LastPacketRx time.Time
}

// Manager manages all BFD sessions.
type Manager struct {
	log      *zap.Logger
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewManager creates a new BFD manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{
		log:      log,
		sessions: make(map[string]*Session),
	}
}

// AddSession creates a new BFD session for a peer.
func (m *Manager) AddSession(peerAddr netip.Addr, txInterval time.Duration, detectMult uint8) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess := &Session{
		PeerAddress:      peerAddr,
		State:            StateDown,
		DetectMultiplier: detectMult,
		TxInterval:       txInterval,
		RxInterval:       txInterval,
	}

	m.sessions[peerAddr.String()] = sess
	m.log.Info("BFD session created",
		zap.String("peer", peerAddr.String()),
		zap.Duration("txInterval", txInterval),
		zap.Uint8("detectMult", detectMult),
	)
	return sess
}
