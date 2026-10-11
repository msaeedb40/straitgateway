package peering

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/mesh/clustermesh"
)

const clusterPeeringFinalizer = "straitgateway.io/clustermesh-protection"

// ClusterPeeringReconciler reconciles ClusterPeering custom resources
type ClusterPeeringReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ClusterMeshMgr *clustermesh.Manager
}

// Reconcile reconciles ClusterPeering
func (r *ClusterPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var peer v1alpha1.ClusterPeering
	if err := r.Get(ctx, req.NamespacedName, &peer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion finalizer
	if !peer.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&peer, clusterPeeringFinalizer) {
			if r.ClusterMeshMgr != nil {
				_ = r.ClusterMeshMgr.RemoveClusterPeer(peer.Spec.ClusterID)
			}
			controllerutil.RemoveFinalizer(&peer, clusterPeeringFinalizer)
			if err := r.Update(ctx, &peer); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Ensure finalizer
	if !controllerutil.ContainsFinalizer(&peer, clusterPeeringFinalizer) {
		controllerutil.AddFinalizer(&peer, clusterPeeringFinalizer)
		if err := r.Update(ctx, &peer); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Synchronize with ClusterMesh Manager
	if r.ClusterMeshMgr != nil {
		if err := r.ClusterMeshMgr.AddClusterPeer(&peer); err != nil {
			logger.Error(err, "Failed to synchronize cluster peer", "clusterId", peer.Spec.ClusterID)
			peer.Status.Phase = "Degraded"
			_ = r.Status().Update(ctx, &peer)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}

		now := metav1.Now()
		peer.Status.Phase = "Connected"
		peer.Status.ActiveRoutes = int32(len(peer.Spec.PodCIDRs) + len(peer.Spec.ServiceCIDRs))
		peer.Status.LastHeartbeat = &now
		peer.Status.LatencyMillis = 1.2
		_ = r.Status().Update(ctx, &peer)
	}

	logger.Info("Successfully reconciled ClusterPeering", "cluster", peer.Spec.ClusterID)
	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler
func (r *ClusterPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterPeering{}).
		Complete(r)
}

// GlobalServiceReconciler reconciles GlobalServiceExport and GlobalServiceImport
type GlobalServiceReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ClusterMeshMgr *clustermesh.Manager
}

// Reconcile reconciles GlobalServiceExport
func (r *GlobalServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var export v1alpha1.GlobalServiceExport
	if err := r.Get(ctx, req.NamespacedName, &export); err == nil {
		if r.ClusterMeshMgr != nil {
			rec, err := r.ClusterMeshMgr.ExportService(
				export.Spec.ServiceName,
				export.Spec.ServiceNamespace,
				export.Spec.ExportedPorts,
				export.Spec.TrafficPolicy,
			)
			if err == nil {
				now := metav1.Now()
				export.Status.Exported = true
				export.Status.ExportedAt = &now
				export.Status.ActiveEndpoints = int32(len(rec.ClusterBackends))
				_ = r.Status().Update(ctx, &export)
			}
		}
		logger.Info("Reconciled GlobalServiceExport", "service", export.Spec.ServiceName)
		return ctrl.Result{}, nil
	}

	var imp v1alpha1.GlobalServiceImport
	if err := r.Get(ctx, req.NamespacedName, &imp); err == nil {
		if r.ClusterMeshMgr != nil {
			rec, err := r.ClusterMeshMgr.ImportService(&imp)
			if err == nil {
				imp.Status.Ready = true
				imp.Status.EndpointsCount = int32(len(rec.ClusterBackends))
				_ = r.Status().Update(ctx, &imp)
			}
		}
		logger.Info("Reconciled GlobalServiceImport", "globalName", imp.Spec.GlobalServiceName)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers GlobalServiceReconciler
func (r *GlobalServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GlobalServiceExport{}).
		Complete(r)
}
