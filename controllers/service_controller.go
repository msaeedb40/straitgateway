// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcmgr "github.com/msaeedb40/straitgateway/service"
)

// ServiceReconciler reconciles Kubernetes Service resources.
// Produces ServiceIR for the dataplane compiler.
type ServiceReconciler struct {
	client.Client
	Log     *zap.Logger
	Manager *svcmgr.Manager
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("service reconcile triggered", zap.String("service", req.NamespacedName.String()))

	// Re-reconcile all services (the manager handles delta detection).
	if _, err := r.Manager.Reconcile(ctx); err != nil {
		r.Log.Error("service reconciliation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Named("service").
		Complete(r)
}
