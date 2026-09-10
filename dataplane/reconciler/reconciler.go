// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package reconciler merges IR objects from all controllers and performs
// delta detection to minimize BPF map writes.
//
// The reconciler is the ONLY input to the compiler. Controllers produce IR;
// the reconciler merges IR; the compiler writes BPF maps.
package reconciler

import (
	"sync"

	"go.uber.org/zap"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
)

// State holds the complete reconciled IR state.
type State struct {
	Services []ir.ServiceIR
	Policies []ir.PolicyIR
	Gateways []ir.GatewayIR
	Transit  []ir.TransitIR
	NAT      []ir.NatIR
}

// Reconciler merges IR from multiple sources and detects deltas.
type Reconciler struct {
	mu       sync.RWMutex
	log      *zap.Logger
	current  State
	previous State
}

// New creates a new Reconciler.
func New(log *zap.Logger) *Reconciler {
	return &Reconciler{log: log}
}

// Update replaces the current state with new IR and returns the delta.
func (r *Reconciler) Update(state State) Delta {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.previous = r.current
	r.current = state

	delta := r.computeDelta()
	r.log.Info("IR reconciliation complete",
		zap.Int("services", len(state.Services)),
		zap.Int("policies", len(state.Policies)),
		zap.Int("gateways", len(state.Gateways)),
		zap.Int("transit", len(state.Transit)),
		zap.Bool("hasChanges", delta.HasChanges()),
	)

	return delta
}

// Current returns the current reconciled state.
func (r *Reconciler) Current() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.current
}

// Delta describes what changed between two reconciliation cycles.
type Delta struct {
	ServicesChanged bool
	PoliciesChanged bool
	GatewaysChanged bool
	TransitChanged  bool
	NATChanged      bool
}

// HasChanges returns true if any IR domain changed.
func (d Delta) HasChanges() bool {
	return d.ServicesChanged || d.PoliciesChanged || d.GatewaysChanged ||
		d.TransitChanged || d.NATChanged
}

func (r *Reconciler) computeDelta() Delta {
	return Delta{
		ServicesChanged: len(r.current.Services) != len(r.previous.Services),
		PoliciesChanged: len(r.current.Policies) != len(r.previous.Policies),
		GatewaysChanged: len(r.current.Gateways) != len(r.previous.Gateways),
		TransitChanged:  len(r.current.Transit) != len(r.previous.Transit),
		NATChanged:      len(r.current.NAT) != len(r.previous.NAT),
	}
}
