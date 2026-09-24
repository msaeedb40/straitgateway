// Package gateway manages Kubernetes Gateway API resources for StraitGateway.
package gateway

import (
	"context"
	"sync"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// Reconciler translates Kubernetes Gateway API resources into StraitD configuration.
type Reconciler struct {
	log    *zap.Logger
	mu     sync.RWMutex
	client *api.Client
}

// NewReconciler creates a new gateway reconciler.
func NewReconciler(log *zap.Logger, client *api.Client) *Reconciler {
	return &Reconciler{
		log:    log,
		client: client,
	}
}

// Reconcile runs a gateway reconciliation pass.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.log.Debug("gateway reconciler pass")
	return nil
}
