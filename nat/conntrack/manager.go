// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package conntrack manages the conntrack LRU BPF map.
// Handles connection tracking for NAT reverse path lookups.
package conntrack

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager manages the conntrack BPF LRU hash map.
type Manager struct {
	log    *zap.Logger
	mu     sync.Mutex
	// gcInterval is the garbage collection interval for expired entries.
	gcInterval time.Duration
}

// New creates a new conntrack Manager.
func New(log *zap.Logger) *Manager {
	return &Manager{
		log:        log,
		gcInterval: 30 * time.Second,
	}
}

// RunGC periodically garbage-collects expired conntrack entries from the BPF map.
func (m *Manager) RunGC(stopCh <-chan struct{}) {
	ticker := time.NewTicker(m.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.gc()
		case <-stopCh:
			return
		}
	}
}

func (m *Manager) gc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Iterate ct_map, delete entries where lifetime has expired.
	m.log.Debug("conntrack GC cycle complete")
}
