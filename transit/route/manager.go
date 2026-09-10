// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package route manages the transit routing table.
// Routes are selected based on topology, segment policy, and route availability.
// Selection priority: Direct Peer → Gateway → Backbone → External BGP.
package route

import (
	"net/netip"
	"sort"
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"go.uber.org/zap"
)

// Protocol indicates how a route was learned.
type Protocol string

const (
	ProtocolWireGuard Protocol = "wireguard"
	ProtocolBGP       Protocol = "bgp"
	ProtocolStatic    Protocol = "static"
	ProtocolKernel    Protocol = "kernel"
)

// TransitRoute represents a route in the transit routing table.
type TransitRoute struct {
	Destination netip.Prefix
	SegmentID   sgtypes.SegmentID
	NextHop     netip.Addr
	GatewayName string
	Protocol    Protocol
	// Metric: lower = higher priority.
	Metric    int
	Encrypted bool
}

// Manager manages the transit routing table.
type Manager struct {
	log    *zap.Logger
	mu     sync.RWMutex
	routes []TransitRoute
}

// New creates a new transit route Manager.
func New(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// Add adds or replaces a route.
func (m *Manager) Add(route TransitRoute) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Remove existing route for same dest+segment.
	m.routes = m.filterOut(route.Destination, route.SegmentID)
	m.routes = append(m.routes, route)
	sort.Slice(m.routes, func(i, j int) bool {
		return m.routes[i].Metric < m.routes[j].Metric
	})
	m.log.Debug("transit route added",
		zap.String("dest", route.Destination.String()),
		zap.Uint32("segment", uint32(route.SegmentID)),
		zap.String("nexthop", route.NextHop.String()),
		zap.String("protocol", string(route.Protocol)),
	)
}

// Remove removes all routes for a destination+segment.
func (m *Manager) Remove(dest netip.Prefix, segID sgtypes.SegmentID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = m.filterOut(dest, segID)
}

// Lookup returns the best route for a destination within a segment.
func (m *Manager) Lookup(dest netip.Addr, segID sgtypes.SegmentID) *TransitRoute {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.routes {
		if r.SegmentID == segID && r.Destination.Contains(dest) {
			return &r
		}
	}
	// Fall back to backbone (segment 0).
	if segID != 0 {
		for _, r := range m.routes {
			if r.SegmentID == 0 && r.Destination.Contains(dest) {
				return &r
			}
		}
	}
	return nil
}

// All returns all routes.
func (m *Manager) All() []TransitRoute {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]TransitRoute, len(m.routes))
	copy(out, m.routes)
	return out
}

func (m *Manager) filterOut(dest netip.Prefix, segID sgtypes.SegmentID) []TransitRoute {
	var out []TransitRoute
	for _, r := range m.routes {
		if !(r.Destination == dest && r.SegmentID == segID) {
			out = append(out, r)
		}
	}
	return out
}
