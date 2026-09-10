// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package transit manages multi-cluster transit gateway state.
// Orchestrates cluster registry, gateway registry, segment manager,
// peer manager, route manager, topology manager, and segment policy engine.
package transit

import (
	"context"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	"github.com/msaeedb40/straitgateway/dataplane/ir"
	"github.com/msaeedb40/straitgateway/transit/cluster"
	gw "github.com/msaeedb40/straitgateway/transit/gateway"
	"github.com/msaeedb40/straitgateway/transit/peer"
	"github.com/msaeedb40/straitgateway/transit/policy"
	"github.com/msaeedb40/straitgateway/transit/route"
	"github.com/msaeedb40/straitgateway/transit/segment"
	"github.com/msaeedb40/straitgateway/transit/topology"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"strconv"
)

// Manager orchestrates the full transit gateway control plane.
type Manager struct {
	client     client.Client
	log        *zap.Logger

	ClusterReg  *cluster.Registry
	GatewayReg  *gw.Registry
	SegmentMgr  *segment.Manager
	PeerMgr     *peer.Manager
	RouteMgr    *route.Manager
	TopologyMgr *topology.Manager
	PolicyEng   *policy.Engine
}

// New creates a new transit Manager with all sub-managers initialized.
func New(c client.Client, log *zap.Logger) *Manager {
	clusterReg := cluster.New()
	gatewayReg := gw.New(log)
	segmentMgr := segment.New(log)
	peerMgr    := peer.New(log)
	routeMgr   := route.New(log)
	topologyMgr := topology.New(log, clusterReg, gatewayReg, peerMgr)
	policyEng  := policy.New(log)

	return &Manager{
		client:      c,
		log:         log,
		ClusterReg:  clusterReg,
		GatewayReg:  gatewayReg,
		SegmentMgr:  segmentMgr,
		PeerMgr:     peerMgr,
		RouteMgr:    routeMgr,
		TopologyMgr: topologyMgr,
		PolicyEng:   policyEng,
	}
}

// Reconcile reads TransitGateway CRDs and returns TransitIR for the dataplane compiler.
func (m *Manager) Reconcile(ctx context.Context) ([]ir.TransitIR, error) {
	var transitList sgv1.TransitGatewayList
	if err := m.client.List(ctx, &transitList); err != nil {
		return nil, err
	}

	var result []ir.TransitIR
	for _, tg := range transitList.Items {
		// Ensure segment exists.
		var clusterID uint64
		if cid, err := strconv.ParseUint(tg.Spec.ClusterID, 10, 32); err == nil {
			clusterID = cid
		}
		segID := sgtypes.SegmentID(clusterID)
		if segID != segment.BackboneSegmentID {
			if _, err := m.SegmentMgr.Get(segID); err != nil {
				_, _ = m.SegmentMgr.Create(segID, tg.Name, "")
			}
		}

		// Reconcile topology.
		m.TopologyMgr.Reconcile(tg.Spec.Topology)

		result = append(result, ir.TransitIR{
			SegmentID: segID,
		})
	}

	m.log.Info("transit reconciliation complete",
		zap.Int("transitGateways", len(transitList.Items)),
		zap.Int("segments", len(m.SegmentMgr.All())),
		zap.Int("peers", len(m.PeerMgr.All())),
	)
	return result, nil
}

// IsAllowed evaluates whether transit traffic is allowed between two segments.
func (m *Manager) IsAllowed(srcSeg, dstSeg sgtypes.SegmentID, proto uint8, dstPort uint16) bool {
	return m.PolicyEng.Evaluate(srcSeg, dstSeg, proto, dstPort) == policy.ActionAllow
}
