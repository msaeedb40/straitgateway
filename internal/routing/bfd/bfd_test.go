package bfd

import (
	"net"
	"testing"
)

func TestBFDFastFailureDetection(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	peerIP := net.ParseIP("192.168.10.5")
	invalidated := false

	session := monitor.AddSession(SessionConfig{
		PeerIP:          peerIP,
		DesiredMinTxMs:  100,
		RequiredMinRxMs: 100,
	}, func(downIP net.IP) {
		if downIP.Equal(peerIP) {
			invalidated = true
		}
	})

	if session.State != StateUp {
		t.Errorf("expected BFD session state Up, got %s", session.State)
	}

	// Simulate link down detection
	err := monitor.TriggerFailure(peerIP)
	if err != nil {
		t.Fatalf("trigger failure failed: %v", err)
	}

	if session.State != StateDown {
		t.Errorf("expected session state Down, got %s", session.State)
	}

	if !invalidated {
		t.Error("expected immediate next-hop invalidation callback on BFD DOWN")
	}
}
