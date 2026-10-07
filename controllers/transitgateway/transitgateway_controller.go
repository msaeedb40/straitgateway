package transitgateway

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/routing"
)

// TransitGatewayReconciler reconciles TransitGateway objects
type TransitGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	FIB    *routing.FIBEngine
}

// Reconcile handles TransitGateway state
func (r *TransitGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var tgw v1alpha1.TransitGateway
	if err := r.Get(ctx, req.NamespacedName, &tgw); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch TransitGateway")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// 1. Sync configured VRFs into the FIB engine
	if r.FIB != nil {
		for _, vrf := range tgw.Spec.VRFs {
			r.FIB.AddVRF(vrf.ID, vrf.Name, vrf.Description)
		}
	}

	// 2. Update status to Ready
	if tgw.Status.Phase != "Ready" {
		tgw.Status.Phase = "Ready"
		if err := r.Status().Update(ctx, &tgw); err != nil {
			logger.Error(err, "failed to update TransitGateway status")
			return ctrl.Result{}, err
		}
	}

	logger.Info("Successfully reconciled TransitGateway", "name", tgw.Name, "asn", tgw.Spec.ASN)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TransitGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TransitGateway{}).
		Complete(r)
}
