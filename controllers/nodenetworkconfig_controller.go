// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
)

// NodeNetworkConfigReconciler reconciles NodeNetworkConfig CRDs.
// These define per-node CIDR allocations, MTU, and tunnel endpoints.
type NodeNetworkConfigReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *NodeNetworkConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("nodenetworkconfig reconcile", zap.String("config", req.NamespacedName.String()))

	var nnc sgv1.NodeNetworkConfig
	if err := r.Get(ctx, req.NamespacedName, &nnc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Info("reconciling NodeNetworkConfig",
		zap.String("node", nnc.Spec.NodeName),
		zap.String("podCIDRv4", nnc.Spec.PodCIDRv4),
	)

	// Apply node-specific configuration to straitgatewayd
	return ctrl.Result{}, nil
}

func (r *NodeNetworkConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sgv1.NodeNetworkConfig{}).
		Named("nodenetworkconfig").
		Complete(r)
}
