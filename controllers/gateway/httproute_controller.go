package gateway

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/straitgateway/straitgateway/internal/gateway"
)

// HTTPRouteReconciler reconciles Gateway API HTTPRoute resources
type HTTPRouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Engine *gateway.HTTPRouteEngine
}

// Reconcile compiles HTTPRoute rules into the fast-path HTTPRouteEngine
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var route gatewayv1.HTTPRoute
	if err := r.Get(ctx, req.NamespacedName, &route); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch HTTPRoute")
			return ctrl.Result{}, err
		}
		// If deleted, remove route from engine
		if r.Engine != nil {
			r.Engine.RemoveRoute(req.Name)
		}
		return ctrl.Result{}, nil
	}

	if r.Engine != nil {
		var compiledRules []*gateway.HTTPRouteRuleConfig

		for i, rule := range route.Spec.Rules {
			cfg := &gateway.HTTPRouteRuleConfig{
				Name:    string(route.Name) + "-rule-" + string(rune('0'+i)),
				Matches: rule.Matches,
			}

			// Convert backend refs
			var totalWeight int32
			for _, b := range rule.BackendRefs {
				weight := int32(1)
				if b.Weight != nil {
					weight = *b.Weight
				}
				port := int32(80)
				if b.Port != nil {
					port = int32(*b.Port)
				}
				dest := gateway.BackendDestination{
					Name:   string(b.Name),
					Port:   port,
					Weight: weight,
				}
				cfg.BackendRefs = append(cfg.BackendRefs, dest)
				totalWeight += weight
			}
			cfg.TotalWeight = totalWeight

			// Parse filters (Headers, Rewrites)
			for _, f := range rule.Filters {
				if f.Type == gatewayv1.HTTPRouteFilterRequestHeaderModifier && f.RequestHeaderModifier != nil {
					cfg.RequestHeaders = f.RequestHeaderModifier
				}
				if f.Type == gatewayv1.HTTPRouteFilterResponseHeaderModifier && f.ResponseHeaderModifier != nil {
					cfg.ResponseHeaders = f.ResponseHeaderModifier
				}
				if f.Type == gatewayv1.HTTPRouteFilterURLRewrite && f.URLRewrite != nil {
					cfg.URLRewrite = f.URLRewrite
				}
			}

			compiledRules = append(compiledRules, cfg)
		}

		// Register for hostnames
		if len(route.Spec.Hostnames) == 0 {
			r.Engine.AddRoute("*", compiledRules)
		} else {
			for _, h := range route.Spec.Hostnames {
				r.Engine.AddRoute(string(h), compiledRules)
			}
		}
	}

	// Update ParentRef status
	if len(route.Spec.ParentRefs) > 0 {
		var parentStatuses []gatewayv1.RouteParentStatus
		for _, p := range route.Spec.ParentRefs {
			st := gatewayv1.RouteParentStatus{
				ParentRef:      p,
				ControllerName: ControllerName,
				Conditions: []metav1.Condition{
					{
						Type:               string(gatewayv1.RouteConditionAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatewayv1.RouteReasonAccepted),
						Message:            "Route accepted by StraitGateway engine",
					},
					{
						Type:               string(gatewayv1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatewayv1.RouteReasonResolvedRefs),
						Message:            "All backend references resolved",
					},
				},
			}
			parentStatuses = append(parentStatuses, st)
		}
		route.Status.Parents = parentStatuses
		_ = r.Status().Update(ctx, &route)
	}

	logger.Info("Reconciled HTTPRoute", "name", route.Name, "namespace", route.Namespace)
	return ctrl.Result{}, nil
}

// SetupWithManager registers the HTTPRouteReconciler
func (r *HTTPRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.HTTPRoute{}).
		Complete(r)
}
