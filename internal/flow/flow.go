// Package flow implements traffic flow inspection for StraitGateway.
//
// It reads from the sg_flow_events ring buffer (eBPF) and exposes
// rich flow telemetry: 5-tuple, connection, session, latency, throughput,
// packet/byte counts, TCP flags, TLS handshake, and DNS events.
package flow

import (
	"context"
	"net"

	"go.uber.org/zap"
)

// FlowKey is the 5-tuple identifying a network flow.
type FlowKey struct {
	SrcAddr  net.IP
	DstAddr  net.IP
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
}

// FlowRecord holds aggregated flow statistics.
type FlowRecord struct {
	Key          FlowKey
	Packets      uint64
	Bytes        uint64
	LastSeenNsec uint64
	SrcIdentity  uint32
	DstIdentity  uint32
	TCPState     uint8
}

// Manager manages flow tracking and ring-buffer consumption.
type Manager struct {
	log *zap.Logger
}

// NewManager creates a new flow Manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// Start starts consuming flow events from the eBPF ring buffer.
func (m *Manager) Start(ctx context.Context) error {
	m.log.Info("flow manager started")
	// TODO:
	//   - Open sg_flow_events ring buffer map
	//   - Consume FlowEvent structs
	//   - Aggregate into FlowRecord cache
	//   - Expose via internal API for sgctl flow / sgpktcap
	return nil
}

// Stop stops the flow manager.
func (m *Manager) Stop() {
	m.log.Info("flow manager stopped")
}

// List returns current active flow records.
func (m *Manager) List() []FlowRecord {
	// TODO: return from in-memory cache
	return nil
}
