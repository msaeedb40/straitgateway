// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package segment manages the lifecycle of transit segments.
//
// Segment 0 is reserved as the backbone. All other segments are isolated by default.
// Segment-to-segment communication requires explicit policy (see transit/policy).
package segment

import (
	"fmt"
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"go.uber.org/zap"
)

const (
	// BackboneSegmentID is the reserved backbone segment (segment 0).
	BackboneSegmentID sgtypes.SegmentID = 0
)

// SegmentInfo describes a transit segment.
type SegmentInfo struct {
	ID          sgtypes.SegmentID
	Name        string
	// CIDR is the address space for this segment (optional, for routing).
	CIDR        string
	// Attachments lists clusters that have joined this segment.
	Attachments []sgtypes.ClusterID
	// BackboneConnected means this segment has connectivity to segment 0.
	BackboneConnected bool
}

// Manager manages the transit segment lifecycle.
type Manager struct {
	log      *zap.Logger
	mu       sync.RWMutex
	segments map[sgtypes.SegmentID]*SegmentInfo
}

// New creates a new segment Manager.
func New(log *zap.Logger) *Manager {
	m := &Manager{
		log:      log,
		segments: make(map[sgtypes.SegmentID]*SegmentInfo),
	}
	// Pre-register backbone.
	m.segments[BackboneSegmentID] = &SegmentInfo{
		ID:   BackboneSegmentID,
		Name: "backbone",
	}
	return m
}

// Create creates a new segment.
func (m *Manager) Create(id sgtypes.SegmentID, name, cidr string) (*SegmentInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id == BackboneSegmentID {
		return nil, fmt.Errorf("cannot create backbone segment (ID 0 is reserved)")
	}
	if _, exists := m.segments[id]; exists {
		return nil, fmt.Errorf("segment %d already exists", id)
	}
	seg := &SegmentInfo{ID: id, Name: name, CIDR: cidr}
	m.segments[id] = seg
	m.log.Info("transit segment created", zap.Uint32("segmentID", uint32(id)), zap.String("name", name))
	return seg, nil
}

// Attach adds a cluster to a segment.
func (m *Manager) Attach(segID sgtypes.SegmentID, clusterID sgtypes.ClusterID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	seg, ok := m.segments[segID]
	if !ok {
		return fmt.Errorf("segment %d not found", segID)
	}
	for _, existing := range seg.Attachments {
		if existing == clusterID {
			return nil // already attached
		}
	}
	seg.Attachments = append(seg.Attachments, clusterID)
	m.log.Info("cluster attached to segment",
		zap.Uint32("segmentID", uint32(segID)),
		zap.Uint32("clusterID", uint32(clusterID)),
	)
	return nil
}

// Detach removes a cluster from a segment.
func (m *Manager) Detach(segID sgtypes.SegmentID, clusterID sgtypes.ClusterID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	seg, ok := m.segments[segID]
	if !ok {
		return fmt.Errorf("segment %d not found", segID)
	}
	newAttachments := seg.Attachments[:0]
	for _, id := range seg.Attachments {
		if id != clusterID {
			newAttachments = append(newAttachments, id)
		}
	}
	seg.Attachments = newAttachments
	return nil
}

// Get returns a segment by ID.
func (m *Manager) Get(id sgtypes.SegmentID) (*SegmentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seg, ok := m.segments[id]
	if !ok {
		return nil, fmt.Errorf("segment %d not found", id)
	}
	return seg, nil
}

// All returns all segments.
func (m *Manager) All() []*SegmentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*SegmentInfo, 0, len(m.segments))
	for _, s := range m.segments {
		out = append(out, s)
	}
	return out
}
