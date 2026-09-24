// Package node manages Kubernetes node discovery and CIDR tracking for StraitGateway.
package node

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// NodeInfo stores discovered node CIDRs without RFC1918 assumptions.
type NodeInfo struct {
	NodeName string
	PodCIDRs []string
}

// Reconciler watches Kubernetes Nodes and maintains per-node CIDRs.
type Reconciler struct {
	log   *zap.Logger
	mu    sync.RWMutex
	nodes map[string]*NodeInfo
}

// NewReconciler creates a new node reconciler.
func NewReconciler(log *zap.Logger) *Reconciler {
	return &Reconciler{
		log:   log,
		nodes: make(map[string]*NodeInfo),
	}
}

// Reconcile runs a node reconciliation pass.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log.Debug("node reconciler pass", zap.Int("trackedNodes", len(r.nodes)))
	return nil
}

// RegisterNode dynamically tracks a node's pod CIDR without hardcoding RFC1918 defaults.
func (r *Reconciler) RegisterNode(name string, cidrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[name] = &NodeInfo{
		NodeName: name,
		PodCIDRs: cidrs,
	}
	r.log.Info("registered node CIDRs", zap.String("node", name), zap.Strings("cidrs", cidrs))
}

// GetNodes returns a snapshot of currently registered nodes.
func (r *Reconciler) GetNodes() map[string]*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make(map[string]*NodeInfo, len(r.nodes))
	for k, v := range r.nodes {
		cp := *v
		res[k] = &cp
	}
	return res
}
