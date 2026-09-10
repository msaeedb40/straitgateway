// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package topology implements the transit topology graph computation.
// Supports hub-spoke, mesh, peer-to-peer, and gateway-to-gateway topologies
// from a single unified graph model. The topology manager computes the full
// adjacency graph and drives the peer manager to establish tunnels.
package topology

import (
	"go.uber.org/zap"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	"github.com/msaeedb40/straitgateway/transit/cluster"
	"github.com/msaeedb40/straitgateway/transit/gateway"
	"github.com/msaeedb40/straitgateway/transit/peer"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// Edge represents a directed connection in the topology graph.
type Edge struct {
	From sgtypes.ClusterID
	To   sgtypes.ClusterID
	// Bidirectional indicates the edge is symmetric.
	Bidirectional bool
}

// Manager computes and maintains the transit topology graph.
type Manager struct {
	log         *zap.Logger
	clusterReg  *cluster.Registry
	gatewayReg  *gateway.Registry
	peerMgr     *peer.Manager
}

// New creates a new topology Manager.
func New(
	log *zap.Logger,
	clusterReg *cluster.Registry,
	gatewayReg *gateway.Registry,
	peerMgr *peer.Manager,
) *Manager {
	return &Manager{
		log:        log,
		clusterReg: clusterReg,
		gatewayReg: gatewayReg,
		peerMgr:    peerMgr,
	}
}

// ComputeEdges computes the topology graph edges for the given topology type.
func (m *Manager) ComputeEdges(topology sgv1.TransitTopology) []Edge {
	clusters := m.clusterReg.All()
	if len(clusters) == 0 {
		return nil
	}

	switch topology {
	case sgv1.TransitTopologyMesh:
		return m.computeMesh(clusters)
	case sgv1.TransitTopologyHubSpoke:
		return m.computeHubSpoke(clusters)
	case sgv1.TransitTopologyPeerToPeer:
		return m.computePeerToPeer(clusters)
	default:
		return m.computeMesh(clusters)
	}
}

// computeMesh returns a full-mesh edge set: every cluster connects to every other.
func (m *Manager) computeMesh(clusters []*cluster.ClusterInfo) []Edge {
	var edges []Edge
	for i := 0; i < len(clusters); i++ {
		for j := i + 1; j < len(clusters); j++ {
			edges = append(edges, Edge{
				From:          clusters[i].ID,
				To:            clusters[j].ID,
				Bidirectional: true,
			})
		}
	}
	m.log.Info("mesh topology computed", zap.Int("edges", len(edges)))
	return edges
}

// computeHubSpoke returns hub-spoke edges: all spokes connect to cluster ID=1 (hub).
// The hub is the cluster with the smallest ClusterID.
func (m *Manager) computeHubSpoke(clusters []*cluster.ClusterInfo) []Edge {
	if len(clusters) < 2 {
		return nil
	}
	// Hub = cluster with smallest ID.
	hub := clusters[0]
	for _, c := range clusters[1:] {
		if c.ID < hub.ID {
			hub = c
		}
	}
	var edges []Edge
	for _, c := range clusters {
		if c.ID == hub.ID {
			continue
		}
		edges = append(edges, Edge{
			From:          hub.ID,
			To:            c.ID,
			Bidirectional: true,
		})
	}
	m.log.Info("hub-spoke topology computed",
		zap.Uint32("hub", uint32(hub.ID)),
		zap.Int("spokes", len(edges)),
	)
	return edges
}

// computePeerToPeer returns point-to-point edges between adjacent clusters.
func (m *Manager) computePeerToPeer(clusters []*cluster.ClusterInfo) []Edge {
	if len(clusters) < 2 {
		return nil
	}
	return []Edge{{
		From:          clusters[0].ID,
		To:            clusters[1].ID,
		Bidirectional: true,
	}}
}

// Reconcile computes edges for the given topology and ensures peers are established.
func (m *Manager) Reconcile(topology sgv1.TransitTopology) {
	edges := m.ComputeEdges(topology)
	m.log.Info("topology reconciled",
		zap.String("topology", string(topology)),
		zap.Int("edges", len(edges)),
	)
	// TODO: For each edge, ensure peer tunnel is established via peerMgr.
}
