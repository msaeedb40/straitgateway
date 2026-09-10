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

// BGPPeerReconciler reconciles BGPPeer CRDs.
type BGPPeerReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *BGPPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("bgp peer reconcile", zap.String("peer", req.NamespacedName.String()))

	var peer sgv1.BGPPeer
	if err := r.Get(ctx, req.NamespacedName, &peer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Info("reconciling BGP peer",
		zap.String("address", peer.Spec.PeerAddress),
		zap.Uint32("asn", peer.Spec.PeerASN),
	)

	// TODO: Feed BGP peer → routing/bgp package → RouteIR
	return ctrl.Result{}, nil
}

func (r *BGPPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sgv1.BGPPeer{}).
		Named("bgppeer").
		Complete(r)
}
