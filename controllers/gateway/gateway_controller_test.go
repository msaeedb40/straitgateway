package gateway

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	internalgw "github.com/straitgateway/straitgateway/internal/gateway"
	"github.com/straitgateway/straitgateway/internal/mesh"
)

func TestGatewayReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = gatewayv1.Install(scheme)

	gw := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prod-gateway",
			Namespace: "default",
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: "straitgateway",
			Listeners: []gatewayv1.Listener{
				{
					Name:     "http",
					Port:     80,
					Protocol: gatewayv1.HTTPProtocolType,
				},
			},
		},
	}

	fakeClient := clientfake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&gatewayv1.Gateway{}).WithObjects(gw).Build()
	proxy := mesh.NewNodeProxy()

	reconciler := &GatewayReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Proxy:  proxy,
	}

	req := ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Namespace: "default",
			Name:      "prod-gateway",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("Gateway reconcile failed: %v", err)
	}

	var updated gatewayv1.Gateway
	_ = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	if len(updated.Status.Conditions) == 0 || updated.Status.Conditions[0].Reason != string(gatewayv1.GatewayReasonProgrammed) {
		t.Errorf("expected Gateway condition to be Programmed, got: %+v", updated.Status.Conditions)
	}
}

func TestHTTPRouteReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = gatewayv1.Install(scheme)

	pathPrefix := "/api/v1"
	pathType := gatewayv1.PathMatchPathPrefix
	weight := int32(100)
	port := gatewayv1.PortNumber(8080)

	route := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-route",
			Namespace: "default",
		},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{"api.straitgateway.io"},
			ParentRefs: []gatewayv1.ParentReference{
				{
					Name: "prod-gateway",
				},
			},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  &pathType,
								Value: &pathPrefix,
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: "api-backend",
									Port: &port,
								},
								Weight: &weight,
							},
						},
					},
				},
			},
		},
	}

	fakeClient := clientfake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&gatewayv1.HTTPRoute{}).WithObjects(route).Build()
	engine := internalgw.NewHTTPRouteEngine()

	reconciler := &HTTPRouteReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Engine: engine,
	}

	req := ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Namespace: "default",
			Name:      "api-route",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("HTTPRoute reconcile failed: %v", err)
	}

	// Verify route was programmed into engine
	testReq := &http.Request{
		Method: "GET",
		Host:   "api.straitgateway.io",
		URL:    &url.URL{Path: "/api/v1/users"},
		Header: make(http.Header),
	}

	res, ok := engine.Match(testReq)
	if !ok || res == nil || res.SelectedTarget.Name != "api-backend" {
		t.Fatalf("expected engine to match api-backend, got %+v", res)
	}

	// Verify status conditions
	var updated gatewayv1.HTTPRoute
	_ = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	if len(updated.Status.Parents) == 0 || len(updated.Status.Parents[0].Conditions) < 2 {
		t.Errorf("expected parent status conditions Accepted and ResolvedRefs")
	}
}
