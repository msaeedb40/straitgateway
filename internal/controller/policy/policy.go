// Package policy manages Kubernetes NetworkPolicy and StraitGatewayPolicy translation for StraitGateway.
package policy

import (
	"context"
	"sync"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// Reconciler watches NetworkPolicies and StraitGatewayPolicies and pushes compiled rules to StraitD.
type Reconciler struct {
	log    *zap.Logger
	mu     sync.RWMutex
	client *api.Client
}

// NewReconciler creates a new policy reconciler.
func NewReconciler(log *zap.Logger, client *api.Client) *Reconciler {
	return &Reconciler{
		log:    log,
		client: client,
	}
}

// Reconcile runs a policy reconciliation pass.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.log.Debug("policy reconciler pass")
	return nil
}
