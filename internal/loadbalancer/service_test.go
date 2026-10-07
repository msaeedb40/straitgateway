package loadbalancer

import (
	"net"
	"testing"
)

func TestServiceManagerUpsertAndGet(t *testing.T) {
	sm := NewServiceManager(nil)

	entry := &ServiceEntry{
		Namespace: "default",
		Name:      "test-service",
		VIP:       net.ParseIP("10.96.0.100"),
		Port:      80,
		Protocol:  6,
		Backends: []Backend{
			{
				ID:     1,
				IP:     net.ParseIP("10.244.1.10"),
				Port:   8080,
				Weight: 100,
				State:  EndpointStateActive,
			},
			{
				ID:     2,
				IP:     net.ParseIP("10.244.2.20"),
				Port:   8080,
				Weight: 100,
				State:  EndpointStateActive,
			},
		},
	}

	if err := sm.UpsertService(entry); err != nil {
		t.Fatalf("failed to upsert service: %v", err)
	}

	got, exists := sm.GetService("default", "test-service", 80)
	if !exists {
		t.Fatal("expected service to exist")
	}

	if len(got.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(got.Backends))
	}

	if got.Maglev == nil {
		t.Error("expected Maglev table to be computed")
	}

	// Delete and verify
	if err := sm.DeleteService("default", "test-service", 80); err != nil {
		t.Fatalf("failed to delete service: %v", err)
	}

	_, exists = sm.GetService("default", "test-service", 80)
	if exists {
		t.Error("expected service to be deleted")
	}
}
