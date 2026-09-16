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

// TLSRouteReconciler reconciles Gateway API TLSRoute resources.
type TLSRouteReconciler struct {
	client client.Client
	log    *zap.Logger
}

// NewTLSRouteReconciler creates a new TLSRoute reconciler.
func NewTLSRouteReconciler(c client.Client, log *zap.Logger) *TLSRouteReconciler {
	return &TLSRouteReconciler{client: c, log: log}
}

// Reconcile reads all TLSRoute resources and returns compiled route rules.
func (r *TLSRouteReconciler) Reconcile(ctx context.Context) ([]ir.RouteRuleIR, error) {
	var routeList gwapiv1a2.TLSRouteList
	if err := r.client.List(ctx, &routeList); err != nil {
		return nil, fmt.Errorf("listing TLSRoutes: %w", err)
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
