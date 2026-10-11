package mesh

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
)

func TestNodeProxyDestinationSelection(t *testing.T) {
	proxy := NewNodeProxy()

	route := &ProxyRoute{
		ServiceRef: "payment-service",
		TrafficSplit: []v1alpha1.TrafficSplitTarget{
			{BackendService: "payment-v1:8080", Weight: 80},
			{BackendService: "payment-v2:8080", Weight: 20},
		},
		Timeout: 5 * time.Second,
	}

	proxy.RegisterRoute(route)

	// Sample multiple selections to verify both destinations are selectable
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		dest, err := proxy.SelectDestination("payment-service")
		if err != nil {
			t.Fatalf("select destination failed: %v", err)
		}
		counts[dest]++
	}

	if counts["payment-v1:8080"] == 0 {
		t.Error("payment-v1:8080 received 0 selections")
	}
	if counts["payment-v2:8080"] == 0 {
		t.Error("payment-v2:8080 received 0 selections")
	}

	// 80/20 split: v1 should have significantly more selections than v2
	if counts["payment-v1:8080"] < counts["payment-v2:8080"] {
		t.Errorf("expected payment-v1 to receive more traffic than payment-v2: %+v", counts)
	}

	// Route inspection tests
	routes := proxy.ListRoutes()
	if len(routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(routes))
	}
	r, found := proxy.GetRoute("payment-service")
	if !found || r == nil {
		t.Error("expected to find route payment-service")
	}

	proxy.RemoveRoute("payment-service")
	if len(proxy.ListRoutes()) != 0 {
		t.Errorf("expected 0 routes after removal, got %d", len(proxy.ListRoutes()))
	}
}

func TestNodeProxyExecuteRequest(t *testing.T) {
	var receivedMeshHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMeshHeader = r.Header.Get("X-StraitGateway-Mesh-Mode")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	proxy := NewNodeProxy()
	host := server.Listener.Addr().String()

	proxy.RegisterRoute(&ProxyRoute{
		ServiceRef: host,
		MTLSMode:   v1alpha1.MTLSStrict,
		Timeout:    5 * time.Second,
	})

	req, _ := http.NewRequest("GET", "http://"+host+"/api/v1/test", nil)
	resp, err := proxy.ExecuteRequest(req, host)
	if err != nil {
		t.Fatalf("execute request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if receivedMeshHeader != string(v1alpha1.MTLSStrict) {
		t.Errorf("expected mesh header Strict, got %s", receivedMeshHeader)
	}
}
