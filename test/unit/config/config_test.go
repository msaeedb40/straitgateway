package config_test

import (
	"testing"

	"github.com/straitgateway/straitgateway/internal/config"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()

	if cfg.Datapath.Mode != "netkit" {
		t.Errorf("default datapath mode = %q, want %q", cfg.Datapath.Mode, "netkit")
	}
	if !cfg.Datapath.EnableXDP {
		t.Error("default datapath.EnableXDP should be true")
	}
	if !cfg.Datapath.EnableTCX {
		t.Error("default datapath.EnableTCX should be true")
	}
	if !cfg.Service.Enabled {
		t.Error("default service.Enabled should be true")
	}
	if !cfg.Policy.Enabled {
		t.Error("default policy.Enabled should be true")
	}
	if cfg.IPAM.Mode != "cluster-pool" {
		t.Errorf("default IPAM mode = %q, want %q", cfg.IPAM.Mode, "cluster-pool")
	}
	if cfg.API.Address == "" {
		t.Error("default API address must not be empty")
	}
}
