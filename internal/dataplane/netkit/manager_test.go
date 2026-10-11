package netkit

import (
	"net"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Mode != ModeL3 {
		t.Errorf("expected default Mode to be ModeL3, got %s", cfg.Mode)
	}
	if cfg.DefaultMTU != 1500 {
		t.Errorf("expected default MTU 1500, got %d", cfg.DefaultMTU)
	}
}

func TestManagerLifecycle(t *testing.T) {
	cfg := DefaultConfig()
	mgr := NewManager(cfg, nil, nil)
	if mgr == nil {
		t.Fatalf("expected non-nil Netkit Manager")
	}

	podIP := net.ParseIP("10.244.0.15")
	primaryName := "sgtest0"
	peerName := "sgtest0_peer"

	// Mock device attachment
	dev := &NetkitDevice{
		PrimaryName: primaryName,
		PeerName:    peerName,
		PrimaryIdx:  10,
		PeerIdx:     11,
		PodIP:       podIP,
		IdentityID:  100,
	}

	mgr.devices[primaryName] = dev

	retrieved, found := mgr.GetDevice(primaryName)
	if !found || retrieved.IdentityID != 100 {
		t.Errorf("expected to find device with IdentityID 100")
	}

	list := mgr.ListDevices()
	if len(list) != 1 {
		t.Errorf("expected 1 device in list, got %d", len(list))
	}

	if err := mgr.DetachPodNetkit(primaryName); err != nil {
		t.Errorf("failed to detach: %v", err)
	}

	if _, found := mgr.GetDevice(primaryName); found {
		t.Errorf("expected device to be removed after detach")
	}

	if err := mgr.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}
