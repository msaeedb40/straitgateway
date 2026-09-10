// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package packets records packet-level drop/action events from the eBPF ring buffer.
// Events are produced by bpf/observability/trace.c and decoded by observability/flow.
// This package re-classifies flow events into packet action records for the dashboard.
package packets

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Action is the per-packet dataplane action.
type Action uint8

const (
	ActionAllow Action = 0
	ActionDrop  Action = 1
	ActionSNAT  Action = 2
	ActionDNAT  Action = 3
)

func (a Action) String() string {
	switch a {
	case ActionAllow: return "ALLOW"
	case ActionDrop:  return "DROP"
	case ActionSNAT:  return "SNAT"
	case ActionDNAT:  return "DNAT"
	default:          return "UNKNOWN"
	}
}

// PacketEvent is a single packet action event from the eBPF dataplane.
type PacketEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	SrcIP       string    `json:"srcIP"`
	DstIP       string    `json:"dstIP"`
	SrcPort     uint16    `json:"srcPort"`
	DstPort     uint16    `json:"dstPort"`
	Protocol    string    `json:"proto"`
	Bytes       uint32    `json:"bytes"`
	Action      Action    `json:"action"`
	DropReason  string    `json:"reason,omitempty"`
	NodeName    string    `json:"nodeName,omitempty"`
	SrcIdentity uint32    `json:"srcIdentity,omitempty"`
	DstIdentity uint32    `json:"dstIdentity,omitempty"`
}

// Recorder buffers packet events from the eBPF ring buffer.
type Recorder struct {
	log      *zap.Logger
	mu       sync.RWMutex
	events   []*PacketEvent
	cap      int
	// Counters for metrics.
	totalDrops  uint64
	totalAllows uint64
	totalNATs   uint64
}

// New creates a new packet Recorder.
func New(log *zap.Logger, capacity int) *Recorder {
	return &Recorder{log: log, cap: capacity}
}

// Record appends a packet event to the ring buffer and updates counters.
func (r *Recorder) Record(ev *PacketEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.events) >= r.cap {
		r.events = r.events[1:]
	}
	r.events = append(r.events, ev)
	switch ev.Action {
	case ActionDrop:  r.totalDrops++
	case ActionAllow: r.totalAllows++
	case ActionSNAT, ActionDNAT: r.totalNATs++
	}
}

// Recent returns the n most recent events (newest-first).
func (r *Recorder) Recent(n int) []*PacketEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if n <= 0 || len(r.events) == 0 {
		return nil
	}
	start := len(r.events) - n
	if start < 0 { start = 0 }
	out := make([]*PacketEvent, len(r.events)-start)
	copy(out, r.events[start:])
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Stats returns aggregate counters.
func (r *Recorder) Stats() (drops, allows, nats uint64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalDrops, r.totalAllows, r.totalNATs
}
