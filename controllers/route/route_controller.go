package route

import (
	"context"
	"net"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/routing"
)

// TransitRouteReconciler reconciles TransitRoute objects
type TransitRouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Router *routing.RouteManager
}

// Reconcile handles TransitRoute synchronization
func (r *TransitRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var tr v1alpha1.TransitRoute
	if err := r.Get(ctx, req.NamespacedName, &tr); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch TransitRoute")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	_, dstNet, err := net.ParseCIDR(tr.Spec.DestinationCIDR)
	if err != nil {
		logger.Error(err, "invalid DestinationCIDR in TransitRoute", "cidr", tr.Spec.DestinationCIDR)
		return ctrl.Result{}, nil
	}

	var nextHopIP net.IP
	if tr.Spec.NextHopIP != "" {
		nextHopIP = net.ParseIP(tr.Spec.NextHopIP)
	}

	if r.Router != nil {
		entry := &routing.RouteEntry{
			Destination: *dstNet,
			VRFID:       tr.Spec.VRF,
			NextHopIP:   nextHopIP,
			Interface:   tr.Spec.OutgoingInterface,
			Precedence:  routing.PrecedenceStaticRoute,
			Metric:      tr.Spec.Metric,
			Communities: tr.Spec.Communities,
		}

		if err := r.Router.SyncRoute(entry); err != nil {
			logger.Error(err, "failed to sync route to FIB/BPF maps")
			return ctrl.Result{}, err
		}
	}

	// Update status
	if !tr.Status.Programmed {
		tr.Status.Programmed = true
		_ = r.Status().Update(ctx, &tr)
	}

	logger.Info("Synchronized TransitRoute", "cidr", tr.Spec.DestinationCIDR, "vrf", tr.Spec.VRF)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TransitRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TransitRoute{}).
		Complete(r)
}
