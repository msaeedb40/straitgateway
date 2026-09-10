// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package gateway manages the registry of transit gateway nodes.
// Gateway nodes are the physical or virtual endpoints that terminate WireGuard
// tunnels and route transit traffic between clusters.
package gateway

import (
	"net/netip"
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"go.uber.org/zap"
)

// GatewayNode represents a transit gateway node.
type GatewayNode struct {
	Name        string
	ClusterID   sgtypes.ClusterID
	// PublicIP is the external IP reachable by other clusters.
	PublicIP    netip.Addr
	// WGPublicKey is this node's WireGuard public key.
	WGPublicKey [32]byte
	// ListenPort is the WireGuard listen port.
	ListenPort  uint16
	// SegmentIDs lists segments this gateway serves.
	SegmentIDs  []sgtypes.SegmentID
	// Role is "hub" or "spoke" for hub-spoke topology.
	Role        string
	// Active indicates this gateway is healthy and active.
	Active      bool
}

// Registry stores all known transit gateway nodes.
type Registry struct {
	log  *zap.Logger
	mu   sync.RWMutex
	gws  map[string]*GatewayNode // keyed by name
}

// New creates a new gateway Registry.
func New(log *zap.Logger) *Registry {
	return &Registry{log: log, gws: make(map[string]*GatewayNode)}
}

// Register adds or updates a gateway node.
func (r *Registry) Register(gw *GatewayNode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gws[gw.Name] = gw
	r.log.Info("transit gateway registered",
		zap.String("name", gw.Name),
		zap.String("ip", gw.PublicIP.String()),
		zap.Uint32("cluster", uint32(gw.ClusterID)),
	)
}

// Deregister removes a gateway node.
func (r *Registry) Deregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.gws, name)
}

// All returns all gateway nodes.
func (r *Registry) All() []*GatewayNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*GatewayNode, 0, len(r.gws))
	for _, g := range r.gws {
		out = append(out, g)
	}
	return out
}

// ForSegment returns all active gateway nodes serving a segment.
func (r *Registry) ForSegment(segID sgtypes.SegmentID) []*GatewayNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*GatewayNode
	for _, g := range r.gws {
		if !g.Active { continue }
		for _, s := range g.SegmentIDs {
			if s == segID { out = append(out, g); break }
		}
	}
	return out
}
