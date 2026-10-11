package gateway

import (
	"net/http"
	"net/url"
	"testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestSmartNICManager(t *testing.T) {
	mgr := NewSmartNICManager()

	// 1. Detect Mellanox device
	devMlx, err := mgr.DetectDevice("mlx5_0")
	if err != nil {
		t.Fatalf("failed to detect device: %v", err)
	}
	if !devMlx.HardwareCapable {
		t.Errorf("expected mlx5_0 to be hardware capable")
	}

	// 2. Attach program in hardware offload mode
	mode, err := mgr.AttachProgram("mlx5_0", "xdp_ingress_prog", ModeHardware)
	if err != nil || mode != ModeHardware {
		t.Errorf("expected ModeHardware, got %s (err: %v)", mode, err)
	}

	// 3. Fallback on virtio device
	devVirtio, _ := mgr.DetectDevice("eth0")
	if devVirtio.HardwareCapable {
		t.Errorf("virtio_net should not be hardware capable")
	}
	modeVirtio, _ := mgr.AttachProgram("eth0", "xdp_ingress_prog", ModeHardware)
	if modeVirtio != ModeNative {
		t.Errorf("expected fallback to ModeNative for non-SmartNIC, got %s", modeVirtio)
	}

	// 4. Detach
	if err := mgr.DetachProgram("mlx5_0"); err != nil {
		t.Errorf("failed to detach: %v", err)
	}
}

func TestHTTPRouteEngine_MatchingAndRewriting(t *testing.T) {
	engine := NewHTTPRouteEngine()

	exactPath := "/api/v1/checkout"
	prefixPath := "/store"
	pathTypePrefix := gatewayv1.PathMatchPathPrefix
	methodPOST := gatewayv1.HTTPMethod("POST")
	headerTypeExact := gatewayv1.HeaderMatchExact
	replacePrefix := "/v2/"

	rules := []*HTTPRouteRuleConfig{
		{
			Name: "canary-checkout",
			Matches: []gatewayv1.HTTPRouteMatch{
				{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  &pathTypePrefix,
						Value: &exactPath,
					},
					Method: &methodPOST,
					Headers: []gatewayv1.HTTPHeaderMatch{
						{
							Type:  &headerTypeExact,
							Name:  "X-Canary",
							Value: "true",
						},
					},
				},
			},
			BackendRefs: []BackendDestination{
				{Name: "checkout-v2", Port: 8080, Weight: 100},
			},
			RequestHeaders: &gatewayv1.HTTPHeaderFilter{
				Set: []gatewayv1.HTTPHeader{
					{Name: "X-Routed-By", Value: "StraitGateway"},
				},
			},
		},
		{
			Name: "default-store",
			Matches: []gatewayv1.HTTPRouteMatch{
				{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  &pathTypePrefix,
						Value: &prefixPath,
					},
				},
			},
			BackendRefs: []BackendDestination{
				{Name: "store-v1", Port: 80, Weight: 100},
			},
			URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
				Path: &gatewayv1.HTTPPathModifier{
					Type:               gatewayv1.PrefixMatchHTTPPathModifier,
					ReplacePrefixMatch: &replacePrefix,
				},
			},
		},
	}

	engine.AddRoute("api.example.com", rules)

	// 1. Test Match on Canary Checkout
	req1 := &http.Request{
		Method: "POST",
		Host:   "api.example.com",
		URL:    &url.URL{Path: "/api/v1/checkout"},
		Header: http.Header{
			"X-Canary": []string{"true"},
		},
	}

	res1, matched1 := engine.Match(req1)
	if !matched1 || res1 == nil {
		t.Fatalf("expected req1 to match canary rule")
	}
	if res1.SelectedTarget.Name != "checkout-v2" {
		t.Errorf("expected checkout-v2 target, got %s", res1.SelectedTarget.Name)
	}
	if req1.Header.Get("X-Routed-By") != "StraitGateway" {
		t.Errorf("expected header filter X-Routed-By: StraitGateway")
	}

	// 2. Test URL Rewrite on Store
	req2 := &http.Request{
		Method: "GET",
		Host:   "api.example.com",
		URL:    &url.URL{Path: "/store/catalog"},
		Header: make(http.Header),
	}

	res2, matched2 := engine.Match(req2)
	if !matched2 || res2 == nil {
		t.Fatalf("expected req2 to match store rule")
	}
	if res2.RewrittenPath != "/v2/catalog" {
		t.Errorf("expected rewritten path /v2/catalog, got %s", res2.RewrittenPath)
	}

	// 3. Test Non-matching request
	req3 := &http.Request{
		Method: "GET",
		Host:   "api.example.com",
		URL:    &url.URL{Path: "/nonexistent"},
		Header: make(http.Header),
	}
	_, matched3 := engine.Match(req3)
	if matched3 {
		t.Errorf("req3 should not have matched")
	}
}

func TestHTTPRouteEngine_WeightedCanaryDistribution(t *testing.T) {
	engine := NewHTTPRouteEngine()

	rules := []*HTTPRouteRuleConfig{
		{
			Name: "weighted-canary",
			BackendRefs: []BackendDestination{
				{Name: "backend-stable", Weight: 80},
				{Name: "backend-canary", Weight: 20},
			},
			TotalWeight: 100,
		},
	}

	engine.AddRoute("*", rules)

	counts := make(map[string]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		req := &http.Request{
			Method: "GET",
			Host:   "test.com",
			URL:    &url.URL{Path: "/"},
			Header: make(http.Header),
		}
		res, ok := engine.Match(req)
		if !ok || res.SelectedTarget == nil {
			t.Fatalf("failed match on iteration %d", i)
		}
		counts[res.SelectedTarget.Name]++
	}

	// 80% stable (expected 700 - 900)
	if counts["backend-stable"] < 700 || counts["backend-stable"] > 900 {
		t.Errorf("unexpected distribution for backend-stable: %d / %d", counts["backend-stable"], iterations)
	}
	// 20% canary (expected 100 - 300)
	if counts["backend-canary"] < 100 || counts["backend-canary"] > 300 {
		t.Errorf("unexpected distribution for backend-canary: %d / %d", counts["backend-canary"], iterations)
	}
}

func TestGRPCRouteEngine(t *testing.T) {
	engine := NewGRPCRouteEngine()

	target := &BackendDestination{Name: "payment-grpc", Port: 50051}
	engine.RegisterMethod("payment.v1.PaymentService", "Authorize", target)

	res, ok := engine.MatchMethod("/payment.v1.PaymentService/Authorize")
	if !ok || res == nil || res.Name != "payment-grpc" {
		t.Errorf("failed to match gRPC method")
	}

	_, ok = engine.MatchMethod("/payment.v1.PaymentService/Unknown")
	if ok {
		t.Errorf("expected no match for unknown gRPC method")
	}
}
