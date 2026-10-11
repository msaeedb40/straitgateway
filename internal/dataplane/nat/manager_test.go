package nat

import (
	"net"
	"testing"
)

func TestNATLifecycle(t *testing.T) {
	mgr := NewManager(nil)
	if mgr == nil {
		t.Fatalf("expected non-nil NAT Manager")
	}

	trans := Translation{
		OrigIP:    net.ParseIP("10.244.0.15"),
		OrigPort:  12345,
		TransIP:   net.ParseIP("192.168.1.50"),
		TransPort: 22345,
		Protocol:  6,
	}

	if err := mgr.SetTranslation(trans); err != nil {
		t.Errorf("unexpected error on SetTranslation: %v", err)
	}

	if err := mgr.DeleteTranslation(trans.OrigIP, trans.OrigPort, trans.Protocol); err != nil {
		t.Errorf("unexpected error on DeleteTranslation: %v", err)
	}
}
