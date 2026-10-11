package conntrack

import (
	"net"
	"testing"
)

func TestBuildFlowKey(t *testing.T) {
	src := net.ParseIP("10.244.0.10")
	dst := net.ParseIP("10.96.0.1")
	key := BuildFlowKey(src, dst, 45678, 443, 6)

	if key.SrcPort != 45678 {
		t.Errorf("expected SrcPort 45678, got %d", key.SrcPort)
	}
	if key.DstPort != 443 {
		t.Errorf("expected DstPort 443, got %d", key.DstPort)
	}
	if key.Proto != 6 {
		t.Errorf("expected Proto 6, got %d", key.Proto)
	}
}

func TestManagerLifecycle(t *testing.T) {
	mgr := NewManager(nil)
	if mgr == nil {
		t.Fatalf("expected non-nil Conntrack Manager")
	}

	state := FlowState{
		SrcIP:      net.ParseIP("10.244.0.10"),
		DstIP:      net.ParseIP("10.96.0.1"),
		SrcPort:    45678,
		DstPort:    443,
		Protocol:   6,
		Packets:    10,
		Bytes:      1500,
		RevSrcIP:   net.ParseIP("10.96.0.1"),
		RevDstIP:   net.ParseIP("10.244.0.10"),
		RevSrcPort: 443,
		RevDstPort: 45678,
	}

	// Without map manager it returns gracefully
	if err := mgr.TrackFlow(state); err != nil {
		t.Errorf("unexpected error on TrackFlow: %v", err)
	}
	if err := mgr.PurgeFlow(state.SrcIP, state.DstIP, state.SrcPort, state.DstPort, state.Protocol); err != nil {
		t.Errorf("unexpected error on PurgeFlow: %v", err)
	}
}
