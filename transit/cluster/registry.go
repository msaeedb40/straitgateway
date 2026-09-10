// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package cluster manages the registry of participating clusters in the transit domain.
package cluster

import (
	"fmt"
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// ClusterInfo holds the runtime state of a registered cluster.
type ClusterInfo struct {
	ID          sgtypes.ClusterID
	Name        string
	APIEndpoint string
	// GatewayNodes lists the IPs of nodes running transit gateway pods.
	GatewayNodes []string
	// PodCIDRs are the pod CIDRs advertised by this cluster.
	PodCIDRs []string
	// ServiceCIDR is the service CIDR of this cluster.
	ServiceCIDR string
	// WGPublicKey is the cluster-level WireGuard public key for this cluster.
	WGPublicKey [32]byte
	// SegmentIDs lists all segments this cluster participates in.
	SegmentIDs []sgtypes.SegmentID
	// Reachable indicates whether this cluster is currently reachable.
	Reachable bool
}

// Registry stores all known clusters in the transit domain.
type Registry struct {
	mu       sync.RWMutex
	clusters map[sgtypes.ClusterID]*ClusterInfo
}

// New creates a new cluster Registry.
func New() *Registry {
	return &Registry{clusters: make(map[sgtypes.ClusterID]*ClusterInfo)}
}

// Register adds or updates a cluster entry.
func (r *Registry) Register(info *ClusterInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clusters[info.ID] = info
}

// Deregister removes a cluster from the registry.
func (r *Registry) Deregister(id sgtypes.ClusterID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clusters, id)
}

// Get returns a cluster by ID.
func (r *Registry) Get(id sgtypes.ClusterID) (*ClusterInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clusters[id]
	if !ok {
		return nil, fmt.Errorf("cluster %d not found", id)
	}
	return c, nil
}

// All returns all registered clusters.
func (r *Registry) All() []*ClusterInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*ClusterInfo, 0, len(r.clusters))
	for _, c := range r.clusters {
		out = append(out, c)
	}
	return out
}

// BySegment returns all clusters participating in a given segment.
func (r *Registry) BySegment(segID sgtypes.SegmentID) []*ClusterInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*ClusterInfo
	for _, c := range r.clusters {
		for _, s := range c.SegmentIDs {
			if s == segID {
				out = append(out, c)
				break
			}
		}
	}
	return out
}
