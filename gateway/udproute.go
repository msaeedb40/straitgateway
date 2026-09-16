// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
)

// UDPRouteReconciler reconciles Gateway API UDPRoute resources.
type UDPRouteReconciler struct {
	client client.Client
	log    *zap.Logger
}

// NewUDPRouteReconciler creates a new UDPRoute reconciler.
func NewUDPRouteReconciler(c client.Client, log *zap.Logger) *UDPRouteReconciler {
	return &UDPRouteReconciler{client: c, log: log}
}

// Reconcile reads all UDPRoute resources and returns compiled route rules.
func (r *UDPRouteReconciler) Reconcile(ctx context.Context) ([]ir.RouteRuleIR, error) {
	var routeList gwapiv1a2.UDPRouteList
	if err := r.client.List(ctx, &routeList); err != nil {
		return nil, fmt.Errorf("listing UDPRoutes: %w", err)
	}

	var rules []ir.RouteRuleIR
	for _, route := range routeList.Items {
		for _, rule := range route.Spec.Rules {
			ruleIR := ir.RouteRuleIR{}
			for _, ref := range rule.BackendRefs {
				be := ir.BackendIR{Weight: 1}
				if ref.Weight != nil {
					be.Weight = uint32(*ref.Weight)
				}
				ruleIR.Backends = append(ruleIR.Backends, be)
			}
			rules = append(rules, ruleIR)
		}
	}
	return rules, nil
}
