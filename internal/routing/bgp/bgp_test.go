package bgp

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/internal/routing"
)

func TestBGPSpeakerLifecycle(t *testing.T) {
	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	speaker := NewSpeaker(65001, net.ParseIP("10.0.0.1"), rm)

	neighborIP := net.ParseIP("192.168.10.2")
	cfg := PeerConfig{
		NeighborIP:      neighborIP,
		RemoteASN:       65002,
		LocalASN:        65001,
		HoldTimeSeconds: 90,
		VRFID:           0,
	}

	peer, err := speaker.AddNeighbor(cfg)
	if err != nil {
		t.Fatalf("failed to add neighbor: %v", err)
	}
	if peer.State != StateEstablished {
		t.Errorf("expected peer state Established, got %s", peer.State)
	}

	// Learn a route from neighbor
	_, prefix, _ := net.ParseCIDR("172.16.0.0/16")
	err = speaker.ReceiveRoute(neighborIP, 0, *prefix, 100)
	if err != nil {
		t.Fatalf("receive route failed: %v", err)
	}

	// Verify route is present in FIB
	matched, err := fib.LongestPrefixMatch(0, net.ParseIP("172.16.5.10"))
	if err != nil {
		t.Fatalf("FIB lookup failed for learned route: %v", err)
	}
	if matched.Destination.String() != "172.16.0.0/16" {
		t.Errorf("expected 172.16.0.0/16, got %s", matched.Destination.String())
	}

	// Remove neighbor and verify route withdrawal
	err = speaker.RemoveNeighbor(neighborIP, 0)
	if err != nil {
		t.Fatalf("remove neighbor failed: %v", err)
	}

	_, err = fib.LongestPrefixMatch(0, net.ParseIP("172.16.5.10"))
	if err == nil {
		t.Error("expected route to be withdrawn after neighbor removal")
	}
}
