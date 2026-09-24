// Package route manages cluster-level route reconciliation for StraitGateway.
package route

import (
	"context"
	"sync"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// Reconciler watches routes and pushes routing table updates to nodes.
type Reconciler struct {
	log    *zap.Logger
	mu     sync.RWMutex
	client *api.Client
}

// NewReconciler creates a new route reconciler.
func NewReconciler(log *zap.Logger, client *api.Client) *Reconciler {
	return &Reconciler{
		log:    log,
		client: client,
	}
}

// Reconcile runs a route reconciliation pass.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.log.Debug("route reconciler pass")
	return nil
}
