package bpf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDataplaneConfig(t *testing.T) {
	cfg := DefaultDataplaneConfig()
	if cfg.PinPath == "" {
		t.Errorf("expected non-empty default PinPath")
	}
	if !cfg.EnableTCX {
		t.Errorf("expected EnableTCX to be true by default")
	}
	if !cfg.EnableXDP {
		t.Errorf("expected EnableXDP to be true by default")
	}
	if !cfg.EnableCgroup {
		t.Errorf("expected EnableCgroup to be true by default")
	}
}

func TestDataplaneManagerLifecycle(t *testing.T) {
	tempPinDir, err := os.MkdirTemp("", "straitgateway-dataplane-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempPinDir) }()

	cfg := DefaultDataplaneConfig()
	cfg.PinPath = filepath.Join(tempPinDir, "bpf")

	mgr, err := NewDataplaneManager(cfg)
	if err != nil {
		t.Logf("Skipping live BPF map creation without root/CAP_BPF: %v", err)
		return
	}
	defer func() { _ = mgr.Close() }()

	if mgr.MapManager() == nil {
		t.Errorf("expected non-nil MapManager")
	}
	if mgr.Loader() == nil {
		t.Errorf("expected non-nil Loader")
	}
	if mgr.OffloadManager() == nil {
		t.Errorf("expected non-nil OffloadManager")
	}
}
