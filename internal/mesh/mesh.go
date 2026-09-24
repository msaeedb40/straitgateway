// Package mesh implements the StraitGateway sidecarless service mesh.
//
// The mesh is implemented through the node-level eBPF datapath.
// No sidecar proxy is required per workload.
package mesh

import (
	"context"

	"go.uber.org/zap"
)

// Manager manages the sidecarless service mesh topology.
type Manager struct {
	log *zap.Logger
}

// NewManager creates a new mesh Manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// Start starts the mesh manager.
func (m *Manager) Start(ctx context.Context) error {
	m.log.Info("mesh manager started")
	// TODO:
	//   - Watch StraitGateway mesh CRDs
	//   - Reconcile sidecarless mesh topology in eBPF
	//   - Enforce mesh policies via policy maps
	//   - Track node and service peers
	return nil
}

// Stop stops the mesh manager.
func (m *Manager) Stop() {
	m.log.Info("mesh manager stopped")
}
