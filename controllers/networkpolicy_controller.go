// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	polcompiler "github.com/msaeedb40/straitgateway/policy"
)

// NetworkPolicyReconciler reconciles standard Kubernetes NetworkPolicy resources.
type NetworkPolicyReconciler struct {
	client.Client
	Log      *zap.Logger
	Compiler *polcompiler.Compiler
}

func (r *NetworkPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("networkpolicy reconcile", zap.String("policy", req.NamespacedName.String()))

	policies, err := r.Compiler.Compile(ctx)
	if err != nil {
		r.Log.Error("policy compilation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	r.Log.Info("policy compilation complete", zap.Int("rules", len(policies)))
	return ctrl.Result{}, nil
}

func (r *NetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.NetworkPolicy{}).
		Named("networkpolicy").
		Complete(r)
}
