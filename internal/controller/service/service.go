// Package service manages Kubernetes Service and EndpointSlice reconciliation for StraitGateway.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// Reconciler watches Services and EndpointSlices and pushes state to nodes.
type Reconciler struct {
	log        *zap.Logger
	mu         sync.RWMutex
	client     *api.Client
	generation int64
	services   map[string]*api.ApplyServiceRequest
}

// NewReconciler creates a new service reconciler.
func NewReconciler(log *zap.Logger, client *api.Client) *Reconciler {
	return &Reconciler{
		log:      log,
		client:   client,
		services: make(map[string]*api.ApplyServiceRequest),
	}
}

// StageService records desired service state for reconciliation.
func (r *Reconciler) StageService(vip string, port uint32, proto string, backends []*api.BackendEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.generation++
	gen := r.generation
	key := fmt.Sprintf("%s:%d/%s", vip, port, proto)

	r.services[key] = &api.ApplyServiceRequest{
		Header: &api.ResourceHeader{
			ResourceUid: key,
			Generation:  gen,
		},
		Key: &api.ServiceKey{
			Addr:  vip,
			Port:  port,
			Proto: proto,
		},
		Backends: backends,
		Flags:    1, // active
	}
}

// Reconcile applies staged service definitions to StraitD.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.log.Debug("service reconciler pass", zap.Int("trackedServices", len(r.services)))
	if r.client == nil {
		return nil
	}

	for _, svc := range r.services {
		if _, err := r.client.ApplyService(ctx, svc); err != nil {
			r.log.Warn("failed to push service to node datapath", zap.Error(err))
		}
	}
	return nil
}
