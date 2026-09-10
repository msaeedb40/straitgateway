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

// TransitGatewayReconciler reconciles TransitGateway, TransitSegment,
// TransitSegmentAttachment, and TransitSegmentRoute resources.
type TransitGatewayReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *TransitGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("transit gateway reconcile", zap.String("transit", req.String()))

	var tg sgv1.TransitGateway
	if err := r.Get(ctx, req.NamespacedName, &tg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Info("reconciling transit gateway",
		zap.String("name", tg.Name),
		zap.String("topology", string(tg.Spec.Topology)),
	)

	// TODO: Reconcile transit segments → TransitIR → Compiler
	return ctrl.Result{}, nil
}

func (r *TransitGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sgv1.TransitGateway{}).
		Named("transitgateway").
		Complete(r)
}
