package mesh

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
)

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		CooldownPeriod:   100 * time.Millisecond,
	})

	if cb.State() != StateClosed {
		t.Errorf("initial state should be Closed, got %s", cb.State())
	}

	// 2 failures: still closed
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed || !cb.Allow() {
		t.Errorf("expected closed after 2 failures")
	}

	// 3rd failure: trips open
	cb.RecordFailure()
	if cb.State() != StateOpen || cb.Allow() {
		t.Errorf("expected Open state and request blocked after 3 failures")
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// Half-open test: Allow should return true once
	if !cb.Allow() {
		t.Errorf("expected Allow() to be true after cooldown in HalfOpen state")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected HalfOpen state, got %s", cb.State())
	}

	// Success restores Closed state
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected closed state after success, got %s", cb.State())
	}
}

func TestCanaryHeaderRouteMatching(t *testing.T) {
	// 1. Setup stable backend
	stableServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "stable-v1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("stable response"))
	}))
	defer stableServer.Close()

	// 2. Setup canary backend
	canaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "canary-v2")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("canary response"))
	}))
	defer canaryServer.Close()

	proxy := NewNodeProxy()
	proxy.RegisterRoute(&ProxyRoute{
		ServiceRef: "app-service",
		TrafficSplit: []v1alpha1.TrafficSplitTarget{
			{BackendService: stableServer.Listener.Addr().String(), Weight: 100},
		},
		CanaryRules: []CanaryHeaderRule{
			{
				HeaderName:    "X-Canary",
				HeaderValue:   "always",
				TargetBackend: canaryServer.Listener.Addr().String(),
			},
		},
	})

	// Case A: Normal request -> stable backend
	reqNormal, _ := http.NewRequest("GET", "http://app-service/items", nil)
	resp, err := proxy.ExecuteRequest(reqNormal, "app-service")
	if err != nil {
		t.Fatalf("normal request failed: %v", err)
	}
	if resp.Header.Get("X-Backend") != "stable-v1" {
		t.Errorf("expected stable-v1, got %s", resp.Header.Get("X-Backend"))
	}
	_ = resp.Body.Close()

	// Case B: Canary header request -> canary backend
	reqCanary, _ := http.NewRequest("GET", "http://app-service/items", nil)
	reqCanary.Header.Set("X-Canary", "always")
	resp, err = proxy.ExecuteRequest(reqCanary, "app-service")
	if err != nil {
		t.Fatalf("canary request failed: %v", err)
	}
	if resp.Header.Get("X-Backend") != "canary-v2" {
		t.Errorf("expected canary-v2, got %s", resp.Header.Get("X-Backend"))
	}
	_ = resp.Body.Close()
}

func TestFaultInjectionAbort(t *testing.T) {
	proxy := NewNodeProxy()
	proxy.RegisterRoute(&ProxyRoute{
		ServiceRef: "faulty-service",
		TrafficSplit: []v1alpha1.TrafficSplitTarget{
			{BackendService: "127.0.0.1:9999", Weight: 100},
		},
		FaultInjection: &FaultInjectionConfig{
			AbortPercentage: 100.0,
			AbortStatusCode: http.StatusTeapot,
		},
	})

	req, _ := http.NewRequest("GET", "http://faulty-service/test", nil)
	_, err := proxy.ExecuteRequest(req, "faulty-service")
	if err == nil {
		t.Fatal("expected fault injection to abort request")
	}
}
