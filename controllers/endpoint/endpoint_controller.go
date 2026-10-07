package endpoint

import (
	"context"
	"net"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/internal/loadbalancer"
)

// EndpointSliceReconciler reconciles EndpointSlice objects
type EndpointSliceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	LB     *loadbalancer.ServiceManager
}

// Reconcile handles EndpointSlice changes
func (r *EndpointSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var slice discoveryv1.EndpointSlice
	if err := r.Get(ctx, req.NamespacedName, &slice); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch EndpointSlice")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	svcName := slice.Labels[discoveryv1.LabelServiceName]
	if svcName == "" {
		return ctrl.Result{}, nil
	}

	// Extract backends from EndpointSlice
	var backends []loadbalancer.Backend
	for i, ep := range slice.Endpoints {
		// Only active/ready endpoints
		if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
			continue
		}
		if len(ep.Addresses) == 0 {
			continue
		}

		ip := net.ParseIP(ep.Addresses[0])
		if ip == nil {
			continue
		}

		nodeName := ""
		if ep.NodeName != nil {
			nodeName = *ep.NodeName
		}

		for _, port := range slice.Ports {
			if port.Port == nil {
				continue
			}
			backends = append(backends, loadbalancer.Backend{
				ID:       uint32(i + 1),
				IP:       ip,
				Port:     uint16(*port.Port),
				Weight:   100,
				State:    loadbalancer.EndpointStateActive,
				NodeName: nodeName,
			})
		}
	}

	for _, port := range slice.Ports {
		if port.Port == nil {
			continue
		}
		// Update existing service entry with backends
		if entry, exists := r.LB.GetService(slice.Namespace, svcName, uint16(*port.Port)); exists {
			entry.Backends = backends
			_ = r.LB.UpsertService(entry)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *EndpointSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		Complete(r)
}
