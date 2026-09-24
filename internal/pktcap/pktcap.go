// Package pktcap implements the StraitGateway packet capture subsystem.
// It reads from the eBPF sg_flow_events ring buffer and exposes captured
// packets via the sgpktcap CLI interface.
package pktcap

import (
	"context"

	"go.uber.org/zap"
)

// Manager manages packet captures on the StraitGateway datapath.
type Manager struct {
	log *zap.Logger
}

// NewManager creates a new pktcap Manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// Start starts the packet capture ring buffer consumer.
func (m *Manager) Start(ctx context.Context) error {
	m.log.Info("pktcap manager started")
	// TODO:
	//   - Open sg_flow_events ring buffer
	//   - Dispatch captured packets to registered capture sessions
	//   - Support BPF filter expressions
	return nil
}

// Stop stops the pktcap manager.
func (m *Manager) Stop() {
	m.log.Info("pktcap manager stopped")
}
