// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/msaeedb40/straitgateway/gateway"
)

// GatewayReconciler reconciles Gateway API GatewayClass, Gateway, and Route resources.
// GatewayClass controller name: straitgateway.io/skgateway
type GatewayReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.With(zap.String("gateway", req.NamespacedName.String()))

	var gw gwapiv1.Gateway
	if err := r.Get(ctx, req.NamespacedName, &gw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Only handle Gateways referencing our GatewayClass.
	if string(gw.Spec.GatewayClassName) != gateway.GatewayClassName {
		return ctrl.Result{}, nil
	}

	log.Info("reconciling gateway",
		zap.String("class", string(gw.Spec.GatewayClassName)),
		zap.Int("listeners", len(gw.Spec.Listeners)),
	)

	// TODO: Compile Gateway listeners → GatewayIR → Compiler
	return ctrl.Result{}, nil
}

func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwapiv1.Gateway{}).
		Named("gateway").
		Complete(r)
}
