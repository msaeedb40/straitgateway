package service

import (
	"context"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/internal/loadbalancer"
)

// ServiceReconciler reconciles a Kubernetes Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	LB     *loadbalancer.ServiceManager
}

// Reconcile handles Service creation, update, and deletion
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var svc corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &svc); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch Service")
			return ctrl.Result{}, err
		}
		// Deleted: cleanup all ports for this service
		return ctrl.Result{}, nil
	}

	// Skip services without ClusterIP or headless services
	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == corev1.ClusterIPNone {
		return ctrl.Result{}, nil
	}

	vip := net.ParseIP(svc.Spec.ClusterIP)
	if vip == nil {
		return ctrl.Result{}, nil
	}

	for _, port := range svc.Spec.Ports {
		proto := uint8(6) // default TCP
		if port.Protocol == corev1.ProtocolUDP {
			proto = 17
		}

		entry := &loadbalancer.ServiceEntry{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			VIP:       vip,
			Port:      uint16(port.Port),
			Protocol:  proto,
		}

		if err := r.LB.UpsertService(entry); err != nil {
			logger.Error(err, "failed to update LB service entry", "service", req.NamespacedName)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
