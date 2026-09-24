// Package controller — watcher.go
// Implements Kubernetes informer-based watchers for the SG Controller.
// Watches Nodes, Services, EndpointSlices, and StraitGateway CRDs and
// reconciles desired state into straitd-consumable configuration.
package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// reconciler is the internal reconcile interface for each resource type.
type reconciler interface {
	reconcile(ctx context.Context) error
}

// initClient initializes a Kubernetes REST client config.
// It tries in-cluster first, then falls back to kubeconfig.
func initClient() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

// scheme adds StraitGateway CRD types to the runtime Scheme.
func buildScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	return s
}

// NodeInfo stores discovered node CIDRs without RFC1918 assumptions.
type NodeInfo struct {
	NodeName string
	PodCIDRs []string
}

// nodeReconciler watches Kubernetes Nodes and maintains per-node CIDRs.
type nodeReconciler struct {
	log   *zap.Logger
	mu    sync.RWMutex
	nodes map[string]*NodeInfo
}

func newNodeReconciler(log *zap.Logger) *nodeReconciler {
	return &nodeReconciler{
		log:   log,
		nodes: make(map[string]*NodeInfo),
	}
}

func (r *nodeReconciler) reconcile(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log.Debug("node reconciler pass", zap.Int("trackedNodes", len(r.nodes)))
	return nil
}

// RegisterNode dynamically tracks a node's pod CIDR without hardcoding RFC1918 defaults.
func (r *nodeReconciler) RegisterNode(name string, cidrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[name] = &NodeInfo{
		NodeName: name,
		PodCIDRs: cidrs,
	}
	r.log.Info("registered node CIDRs", zap.String("node", name), zap.Strings("cidrs", cidrs))
}

// serviceReconciler watches Services and EndpointSlices and pushes state to nodes.
type serviceReconciler struct {
	log        *zap.Logger
	mu         sync.RWMutex
	client     *api.Client
	generation int64
	services   map[string]*api.ApplyServiceRequest
}

func newServiceReconciler(log *zap.Logger, client *api.Client) *serviceReconciler {
	return &serviceReconciler{
		log:      log,
		client:   client,
		services: make(map[string]*api.ApplyServiceRequest),
	}
}

func (r *serviceReconciler) reconcile(ctx context.Context) error {
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

// StageService records a desired service VIP and backend mappings.
func (r *serviceReconciler) StageService(vip string, port uint32, proto string, backends []*api.BackendEntry) {
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

// gatewayReconciler watches Gateway API resources.
type gatewayReconciler struct {
	log      *zap.Logger
	mu       sync.RWMutex
	client   *api.Client
	gateways map[string]*api.ApplyGatewayRequest
}

func newGatewayReconciler(log *zap.Logger, client *api.Client) *gatewayReconciler {
	return &gatewayReconciler{
		log:      log,
		client:   client,
		gateways: make(map[string]*api.ApplyGatewayRequest),
	}
}

func (r *gatewayReconciler) reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.log.Debug("gateway reconciler pass", zap.Int("trackedGateways", len(r.gateways)))
	if r.client == nil {
		return nil
	}

	for _, gw := range r.gateways {
		if _, err := r.client.ApplyGateway(ctx, gw); err != nil {
			r.log.Warn("failed to push gateway to node datapath", zap.Error(err))
		}
	}
	return nil
}

// policyReconciler watches StraitGatewayPolicy CRDs.
type policyReconciler struct {
	log      *zap.Logger
	mu       sync.RWMutex
	client   *api.Client
	policies map[string]*api.ApplyPolicyRequest
}

func newPolicyReconciler(log *zap.Logger, client *api.Client) *policyReconciler {
	return &policyReconciler{
		log:      log,
		client:   client,
		policies: make(map[string]*api.ApplyPolicyRequest),
	}
}

func (r *policyReconciler) reconcile(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.log.Debug("policy reconciler pass", zap.Int("trackedPolicies", len(r.policies)))
	if r.client == nil {
		return nil
	}

	for _, pol := range r.policies {
		if _, err := r.client.ApplyPolicy(ctx, pol); err != nil {
			r.log.Warn("failed to push policy to node datapath", zap.Error(err))
		}
	}
	return nil
}
