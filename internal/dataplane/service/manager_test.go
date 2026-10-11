package service

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/internal/loadbalancer"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CgroupPath != "/sys/fs/cgroup" {
		t.Errorf("expected default cgroup path /sys/fs/cgroup, got %s", cfg.CgroupPath)
	}
}

func TestServiceManagerLifecycle(t *testing.T) {
	cfg := DefaultConfig()
	mgr := NewManager(cfg, nil, nil)
	if mgr == nil {
		t.Fatalf("expected non-nil Service Dataplane Manager")
	}

	vip := net.ParseIP("10.96.0.10")
	backends := []loadbalancer.Backend{
		{
			IP:     net.ParseIP("10.244.0.5"),
			Port:   8080,
			Weight: 100,
			State:  loadbalancer.EndpointStateActive,
		},
		{
			IP:     net.ParseIP("10.244.1.7"),
			Port:   8080,
			Weight: 100,
			State:  loadbalancer.EndpointStateActive,
		},
	}

	err := mgr.RegisterService("default", "web-svc", vip, 80, 6, backends)
	if err != nil {
		t.Fatalf("failed to register service: %v", err)
	}

	selected, err := mgr.ServiceManager().SelectBackend("default", "web-svc", 80, 12345)
	if err != nil {
		t.Fatalf("failed to select backend: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected selected backend to be non-nil")
	}

	if err := mgr.UnregisterService("default", "web-svc", 80); err != nil {
		t.Errorf("failed to unregister service: %v", err)
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}
