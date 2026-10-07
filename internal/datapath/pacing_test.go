package datapath

import (
	"testing"
)

func TestPacingManagerRateComputation(t *testing.T) {
	pm := NewPacingManager(PacingConfig{
		DefaultLinkRateBps: 10 * 1000 * 1000 * 1000, // 10 Gbps
		TargetRTTUs:        200,                     // 200 us
		MinPacingRateBps:   100 * 1000 * 1000,       // 100 Mbps
		Algorithm:          PacingEDT,
	})

	// 1. Test standard flow rate computation with 15% headroom
	rate := pm.ComputePacingRate(200, 10*1000*1000*1000)
	expected := uint64(11.5 * 1000 * 1000 * 1000) // 11.5 Gbps
	if rate != expected {
		t.Errorf("expected pacing rate %d, got %d", expected, rate)
	}

	// 2. Test flow registration and BDP calculation
	flow := pm.RegisterFlow("tcp:10.244.1.5:45214->10.96.0.1:443", 200, 10*1000*1000*1000)
	if flow == nil || flow.CurrentRateBps != expected {
		t.Fatalf("unexpected registered flow: %+v", flow)
	}
	if flow.BDPBytes == 0 {
		t.Errorf("BDP bytes should be non-zero")
	}

	// 3. Test zero-window probe backoff
	err := pm.RecordZeroWindow("tcp:10.244.1.5:45214->10.96.0.1:443")
	if err != nil {
		t.Fatalf("unexpected error recording zero-window: %v", err)
	}
	backedOffFlow, exists := pm.GetFlow("tcp:10.244.1.5:45214->10.96.0.1:443")
	if !exists || backedOffFlow.ZeroWindowProbes != 1 {
		t.Errorf("expected 1 zero-window probe")
	}
	if backedOffFlow.CurrentRateBps != expected/2 {
		t.Errorf("expected 50%% rate backoff to %d, got %d", expected/2, backedOffFlow.CurrentRateBps)
	}
}
