// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sgv1alpha1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
)

// IPAMReconciler watches Node resources and dynamically manages per-node CIDR allocations.
// It discovers podCIDR from node.spec.podCIDR and creates/updates NodeNetworkConfig CRDs.
type IPAMReconciler struct {
	client.Client
	Log *zap.Logger
}

func (r *IPAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("IPAM reconcile", zap.String("node", req.String()))

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	podCIDR := node.Spec.PodCIDR
	if podCIDR == "" && len(node.Spec.PodCIDRs) > 0 {
		podCIDR = node.Spec.PodCIDRs[0]
	}

	r.Log.Info("IPAM: node CIDR dynamically discovered",
		zap.String("node", node.Name),
		zap.String("podCIDR", podCIDR),
	)

	// Create or update NodeNetworkConfig CRD for this node.
	nnConfigName := types.NamespacedName{
		Name: node.Name,
	}

	var nnConfig sgv1alpha1.NodeNetworkConfig
	err := r.Get(ctx, nnConfigName, &nnConfig)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	if err != nil {
		// Does not exist, create it
		nnConfig = sgv1alpha1.NodeNetworkConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: node.Name,
			},
			Spec: sgv1alpha1.NodeNetworkConfigSpec{
				NodeName:  node.Name,
				PodCIDRv4: podCIDR,
				Overlay:   sgv1alpha1.OverlayModeNative,
				IPAM:      sgv1alpha1.IPAMModePerNode,
			},
		}
		if len(node.Spec.PodCIDRs) > 1 {
			nnConfig.Spec.PodCIDRv6 = node.Spec.PodCIDRs[1]
		}
		if err := r.Create(ctx, &nnConfig); err != nil {
			r.Log.Error("failed to create NodeNetworkConfig", zap.String("node", node.Name), zap.Error(err))
			return ctrl.Result{}, err
		}
		r.Log.Info("created NodeNetworkConfig", zap.String("node", node.Name), zap.String("podCIDR", podCIDR))
	} else {
		// Update if CIDR changed
		if nnConfig.Spec.PodCIDRv4 != podCIDR {
			nnConfig.Spec.PodCIDRv4 = podCIDR
			if len(node.Spec.PodCIDRs) > 1 {
				nnConfig.Spec.PodCIDRv6 = node.Spec.PodCIDRs[1]
			}
			if err := r.Update(ctx, &nnConfig); err != nil {
				r.Log.Error("failed to update NodeNetworkConfig", zap.String("node", node.Name), zap.Error(err))
				return ctrl.Result{}, err
			}
			r.Log.Info("updated NodeNetworkConfig", zap.String("node", node.Name), zap.String("podCIDR", podCIDR))
		}
	}

	return ctrl.Result{}, nil
}

func (r *IPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Named("ipam").
		Complete(r)
}
