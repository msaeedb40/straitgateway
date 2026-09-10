// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package policy implements the transit segment connectivity policy engine.
//
// Default action: DENY. Backbone connectivity (segment 0) does NOT automatically
// grant inter-segment connectivity. All cross-segment traffic requires explicit allow.
package policy

import (
	"sync"

	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"go.uber.org/zap"
)

// Action is the policy action for segment-to-segment traffic.
type Action uint8

const (
	ActionDeny  Action = 0
	ActionAllow Action = 1
)

// SegmentPolicy defines the connectivity policy between two segments.
type SegmentPolicy struct {
	SrcSegment sgtypes.SegmentID
	DstSegment sgtypes.SegmentID
	Action     Action
	// Priority: 0=highest.
	Priority int
	// Protocols restricts to specific L4 protocols (nil=all).
	Protocols []uint8
	// Ports restricts to specific destination ports (nil=all).
	Ports []uint16
}

// Engine evaluates transit segment connectivity policies.
type Engine struct {
	log      *zap.Logger
	mu       sync.RWMutex
	policies []SegmentPolicy
}

// New creates a new transit policy Engine.
func New(log *zap.Logger) *Engine {
	return &Engine{log: log}
}

// AddPolicy adds a segment connectivity policy.
func (e *Engine) AddPolicy(p SegmentPolicy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies = append(e.policies, p)
	e.log.Info("transit segment policy added",
		zap.Uint32("src", uint32(p.SrcSegment)),
		zap.Uint32("dst", uint32(p.DstSegment)),
		zap.Uint8("action", uint8(p.Action)),
	)
}

// Evaluate returns whether traffic from srcSegment to dstSegment is allowed.
// Same segment is always allowed. Default is deny.
func (e *Engine) Evaluate(srcSeg, dstSeg sgtypes.SegmentID, protocol uint8, dstPort uint16) Action {
	// Same segment: always allow.
	if srcSeg == dstSeg {
		return ActionAllow
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// First-match wins (lowest priority number first).
	for _, p := range e.policies {
		if p.SrcSegment != srcSeg && p.SrcSegment != 0 {
			continue
		}
		if p.DstSegment != dstSeg && p.DstSegment != 0 {
			continue
		}
		if len(p.Protocols) > 0 {
			matched := false
			for _, proto := range p.Protocols {
				if proto == protocol { matched = true; break }
			}
			if !matched { continue }
		}
		if len(p.Ports) > 0 {
			matched := false
			for _, port := range p.Ports {
				if port == dstPort { matched = true; break }
			}
			if !matched { continue }
		}
		return p.Action
	}

	// Default deny.
	return ActionDeny
}
