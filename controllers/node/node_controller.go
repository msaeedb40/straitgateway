package node

import (
	"context"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/internal/routing"
)

// NodeReconciler reconciles Kubernetes Node objects for cluster-wide mesh topology
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Router *routing.RouteManager
}

// Reconcile handles Node PodCIDR discovery
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch Node")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Discover node IP
	var nodeIP net.IP
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP = net.ParseIP(addr.Address)
			break
		}
	}

	// If node has assigned PodCIDR and nodeIP, install cluster route
	if node.Spec.PodCIDR != "" && nodeIP != nil && r.Router != nil {
		_, podCIDR, err := net.ParseCIDR(node.Spec.PodCIDR)
		if err == nil {
			entry := &routing.RouteEntry{
				Destination: *podCIDR,
				VRFID:       0,
				NextHopIP:   nodeIP,
				Precedence:  routing.PrecedenceClusterRoute,
				Metric:      50,
			}
			_ = r.Router.SyncRoute(entry)
			logger.Info("Installed cluster route for node", "node", node.Name, "cidr", node.Spec.PodCIDR, "nextHop", nodeIP)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
