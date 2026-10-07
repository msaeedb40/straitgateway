package loader

import (
	"testing"
)

func TestOffloadManagerDetectCapabilities(t *testing.T) {
	om := NewOffloadManager()

	// 1. Mock Mellanox driver
	om.SetMockDriver("lo", "mlx5_core")
	caps, err := om.DetectCapabilities("lo")
	if err != nil {
		t.Fatalf("unexpected error detecting lo with mlx5: %v", err)
	}
	if caps.Vendor != VendorMellanox {
		t.Errorf("expected vendor %s, got %s", VendorMellanox, caps.Vendor)
	}
	if !caps.SupportsOffload {
		t.Errorf("expected supportsOffload to be true for Mellanox")
	}
	if caps.LineRateGbps != 100 {
		t.Errorf("expected 100Gbps line rate, got %d", caps.LineRateGbps)
	}

	// 2. Mock Netronome driver
	om.SetMockDriver("lo", "nfp")
	caps, err = om.DetectCapabilities("lo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if caps.Vendor != VendorNetronome || !caps.SupportsOffload {
		t.Errorf("expected Netronome with offload support, got %+v", caps)
	}

	// 3. Mock Standard veth driver
	om.SetMockDriver("lo", "veth")
	caps, err = om.DetectCapabilities("lo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if caps.Vendor != VendorGeneric {
		t.Errorf("expected Generic vendor for veth, got %s", caps.Vendor)
	}
	if caps.SupportsOffload {
		t.Errorf("generic veth should not support hardware offload")
	}
}

func TestOffloadManagerAttachFallback(t *testing.T) {
	om := NewOffloadManager()
	om.SetMockDriver("lo", "veth")

	// Test attaching without live BPF prog: should gracefully report mode
	mode, err := om.AttachProgram("lo", nil, XDPModeGeneric)
	if err != nil {
		t.Fatalf("unexpected error during attach generic: %v", err)
	}
	if mode != XDPModeGeneric {
		t.Errorf("expected mode %s, got %s", XDPModeGeneric, mode)
	}

	// Detach
	if err := om.DetachProgram("lo"); err != nil {
		t.Fatalf("failed to detach: %v", err)
	}
}
