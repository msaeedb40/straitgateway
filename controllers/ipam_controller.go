// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IPAMReconciler watches Node resources and manages per-node CIDR allocations.
// It discovers podCIDR from node.spec.podCIDR and creates NodeNetworkConfig CRDs.
type IPAMReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *IPAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("IPAM reconcile", zap.String("node", req.NamespacedName.String()))

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Info("IPAM: node CIDR discovered",
		zap.String("node", node.Name),
		zap.String("podCIDR", node.Spec.PodCIDR),
	)

	// TODO: Create/update NodeNetworkConfig CRD with podCIDR range
	return ctrl.Result{}, nil
}

func (r *IPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Named("ipam").
		Complete(r)
}
