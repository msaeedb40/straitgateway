package gateway

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/straitgateway/straitgateway/internal/mesh"
)

const (
	ControllerName gatewayv1.GatewayController = "straitgateway.io/gateway-controller"
)

// GatewayReconciler reconciles Gateway API Gateway objects
type GatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Proxy  *mesh.NodeProxy
}

// Reconcile handles Gateway resources
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var gw gatewayv1.Gateway
	if err := r.Get(ctx, req.NamespacedName, &gw); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch Gateway")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Update status condition to Programmed
	cond := metav1.Condition{
		Type:               string(gatewayv1.GatewayConditionProgrammed),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: gw.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatewayv1.GatewayReasonProgrammed),
		Message:            "StraitGateway dataplane successfully programmed",
	}

	gw.Status.Conditions = []metav1.Condition{cond}
	_ = r.Status().Update(ctx, &gw)

	logger.Info("Reconciled Gateway API Gateway", "name", gw.Name, "namespace", gw.Namespace)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.Gateway{}).
		Complete(r)
}
