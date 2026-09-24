package config

import (
	"testing"
)

func TestConfigValidation_DefaultIsValid(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Defaults() should be valid, got: %v", err)
	}
}

func TestConfigValidation_InvalidDatapathMode(t *testing.T) {
	cfg := Defaults()
	cfg.Datapath.Mode = "invalid-mode"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid datapath mode, got nil")
	}
}

func TestConfigValidation_InvalidCIDR(t *testing.T) {
	cfg := Defaults()
	cfg.IPAM.PodCIDRs = []string{"999.999.999.999/24"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid CIDR, got nil")
	}
}

func TestConfigValidation_InvalidTopology(t *testing.T) {
	cfg := Defaults()
	cfg.Transit.Enabled = true
	cfg.Transit.Topology = "star"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid transit topology, got nil")
	}
}
