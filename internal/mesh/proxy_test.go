package mesh

import (
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
}
