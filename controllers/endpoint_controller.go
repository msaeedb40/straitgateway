// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	discoveryv1 "k8s.io/api/discovery/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcmgr "github.com/msaeedb40/straitgateway/service"
)

// EndpointSliceReconciler watches EndpointSlice changes and triggers service
// reconciliation to update backend lists.
type EndpointSliceReconciler struct {
	client.Client
	Log     *zap.Logger
	Manager *svcmgr.Manager
}

func (r *EndpointSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("endpointslice reconcile triggered", zap.String("eps", req.String()))

	if _, err := r.Manager.Reconcile(ctx); err != nil {
		r.Log.Error("endpoint reconciliation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EndpointSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		Named("endpointslice").
		Complete(r)
}
