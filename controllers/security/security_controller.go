package security

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/mesh"
)

// SecurityPolicyReconciler reconciles SecurityPolicy objects
type SecurityPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Compiler *mesh.PolicyCompiler
	Identity *mesh.IdentityEngine
}

// Reconcile compiles SecurityPolicy rules into BPF maps
func (r *SecurityPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var policy v1alpha1.SecurityPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch SecurityPolicy")
			return ctrl.Result{}, err
		}
		// Deleted: remove compiled rules
		_ = r.Compiler.RemovePolicy(req.Namespace, req.Name)
		return ctrl.Result{}, nil
	}

	// Resolve target identity ID
	targetIdent := r.Identity.GetOrCreateIdentity(policy.Namespace, "default", nil)
	if policy.Spec.Selector != nil && len(policy.Spec.Selector.MatchLabels) > 0 {
		targetIdent = r.Identity.GetOrCreateIdentity(policy.Namespace, "default", policy.Spec.Selector.MatchLabels)
	}

	// Compile and install
	rules, err := r.Compiler.CompileAndInstall(&policy, targetIdent.ID)
	if err != nil {
		logger.Error(err, "failed to compile security policy", "policy", req.NamespacedName)
		return ctrl.Result{}, err
	}

	logger.Info("Compiled security policy rules", "count", len(rules), "targetID", targetIdent.ID)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SecurityPolicy{}).
		Complete(r)
}
