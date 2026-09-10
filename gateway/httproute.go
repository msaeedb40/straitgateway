// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
)

// HTTPRouteReconciler reconciles HTTPRoute resources.
type HTTPRouteReconciler struct {
	client client.Client
	log    *zap.Logger
}

// NewHTTPRouteReconciler creates a new HTTPRoute reconciler.
func NewHTTPRouteReconciler(c client.Client, log *zap.Logger) *HTTPRouteReconciler {
	return &HTTPRouteReconciler{client: c, log: log}
}

// Reconcile reads all HTTPRoute resources and returns compiled route rules.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context) ([]ir.RouteRuleIR, error) {
	var routeList gwapiv1.HTTPRouteList
	if err := r.client.List(ctx, &routeList); err != nil {
		return nil, fmt.Errorf("listing HTTPRoutes: %w", err)
	}

	var rules []ir.RouteRuleIR
	for _, route := range routeList.Items {
		for _, rule := range route.Spec.Rules {
			ruleIR := ir.RouteRuleIR{}

			// Extract path prefix from matches.
			if len(rule.Matches) > 0 && rule.Matches[0].Path != nil {
				if rule.Matches[0].Path.Value != nil {
					ruleIR.PathPrefix = *rule.Matches[0].Path.Value
				}
			}

			// Extract header matches.
			ruleIR.Headers = make(map[string]string)
			if len(rule.Matches) > 0 {
				for _, hdr := range rule.Matches[0].Headers {
					ruleIR.Headers[string(hdr.Name)] = hdr.Value
				}
			}

			// Extract backend references.
			for _, ref := range rule.BackendRefs {
				be := ir.BackendIR{
					Weight: 1,
				}
				if ref.Weight != nil {
					be.Weight = uint32(*ref.Weight)
				}
				ruleIR.Backends = append(ruleIR.Backends, be)
			}

			rules = append(rules, ruleIR)
		}
	}

	r.log.Info("HTTPRoute reconciliation complete", zap.Int("rules", len(rules)))
	return rules, nil
}
