// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	polcompiler "github.com/msaeedb40/straitgateway/policy"
)

// StraitNetworkPolicyReconciler reconciles StraitNetworkPolicy CRDs.
type StraitNetworkPolicyReconciler struct {
	client.Client
	Log      *zap.Logger
	Compiler *polcompiler.Compiler
}

func (r *StraitNetworkPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("straitnetworkpolicy reconcile", zap.String("policy", req.String()))

	policies, err := r.Compiler.Compile(ctx)
	if err != nil {
		r.Log.Error("policy compilation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	r.Log.Info("straitnetworkpolicy compiled", zap.Int("rules", len(policies)))
	return ctrl.Result{}, nil
}

func (r *StraitNetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sgv1.StraitNetworkPolicy{}).
		Named("straitnetworkpolicy").
		Complete(r)
}
