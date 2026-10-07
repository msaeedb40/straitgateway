package agent

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	status := hc.Check()

	if status.KernelRelease == "" {
		t.Error("expected non-empty kernel release string")
	}

	summary := status.FormatStatus()
	if summary == "" {
		t.Error("expected non-empty format summary")
	}
}

func TestAgentManagerLifecycle(t *testing.T) {
	cfg := Config{
		NodeName:       "test-node",
		NodeIP:         net.ParseIP("192.168.1.50"),
		LocalASN:       65001,
		HealthInterval: 50 * time.Millisecond,
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create agent manager: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run agent in background and verify graceful exit on context cancellation
	errCh := make(chan error, 1)
	go func() {
		errCh <- mgr.Start(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("agent exited with unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("agent failed to exit after timeout")
	}
}
