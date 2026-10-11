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

func TestBFDHeartbeatAndTimeout(t *testing.T) {
	monitor := NewMonitor()
	defer monitor.Close()

	peerIP := net.ParseIP("192.168.20.10")
	downCount := 0

	session := monitor.AddSession(SessionConfig{
		PeerIP:           peerIP,
		DesiredMinTxMs:   10,
		RequiredMinRxMs:  10,
		DetectMultiplier: 2, // 20ms timeout
	}, func(downIP net.IP) {
		downCount++
	})

	// Heartbeat updates LastRx
	ok := monitor.ReceiveHeartbeat(peerIP)
	if !ok {
		t.Error("expected ReceiveHeartbeat to return true")
	}

	// Verify listing sessions
	sessions := monitor.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}

	// No timeout right away
	failed := monitor.CheckTimeouts()
	if len(failed) != 0 {
		t.Errorf("expected 0 timed-out peers, got %d", len(failed))
	}

	// Manually set LastRx into the past to trigger timeout
	session.LastRx = session.LastRx.Add(-50 * 1000 * 1000) // 50ms in past

	failed = monitor.CheckTimeouts()
	if len(failed) != 1 {
		t.Errorf("expected 1 timed-out peer, got %d", len(failed))
	}
	if session.State != StateDown {
		t.Errorf("expected state Down, got %s", session.State)
	}
	if downCount != 1 {
		t.Errorf("expected down callback triggered once, got %d", downCount)
	}

	// Remove session
	monitor.RemoveSession(peerIP)
	if len(monitor.ListSessions()) != 0 {
		t.Errorf("expected 0 sessions after removal, got %d", len(monitor.ListSessions()))
	}
}
