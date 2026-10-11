package bgp

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
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

func TestBGPWithBFDFastFailover(t *testing.T) {
	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	speaker := NewSpeaker(65001, net.ParseIP("10.0.0.1"), rm)

	bfdMon := bfd.NewMonitor()
	defer bfdMon.Close()
	speaker.SetBFDMonitor(bfdMon)

	neighborIP := net.ParseIP("192.168.30.2")
	cfg := PeerConfig{
		NeighborIP:      neighborIP,
		RemoteASN:       65003,
		LocalASN:        65001,
		EnableBFD:       true,
		HoldTimeSeconds: 90,
		VRFID:           0,
	}

	_, err := speaker.AddNeighbor(cfg)
	if err != nil {
		t.Fatalf("failed to add neighbor with BFD: %v", err)
	}

	// Check BFD session was registered
	session, exists := bfdMon.GetSession(neighborIP)
	if !exists || session == nil {
		t.Fatal("expected BFD session to be registered for neighbor")
	}

	// Receive 2 routes from this neighbor
	_, p1, _ := net.ParseCIDR("10.50.0.0/16")
	_, p2, _ := net.ParseCIDR("10.60.0.0/16")
	_ = speaker.ReceiveRoute(neighborIP, 0, *p1, 50)
	_ = speaker.ReceiveRoute(neighborIP, 0, *p2, 50)

	// Verify both routes are reachable
	if _, err := fib.LongestPrefixMatch(0, net.ParseIP("10.50.1.1")); err != nil {
		t.Fatalf("expected route 1 reachable: %v", err)
	}
	if _, err := fib.LongestPrefixMatch(0, net.ParseIP("10.60.1.1")); err != nil {
		t.Fatalf("expected route 2 reachable: %v", err)
	}

	// Simulate BFD trigger failure (sub-second link fail)
	err = bfdMon.TriggerFailure(neighborIP)
	if err != nil {
		t.Fatalf("failed to trigger failure: %v", err)
	}

	// Verify session degraded
	peer, _ := speaker.GetPeer(neighborIP, 0)
	if peer.State != StateActive {
		t.Errorf("expected peer state Active/recovering, got %s", peer.State)
	}

	// Crucial: verify learned routes were IMMEDIATELY withdrawn from the FIB!
	if _, err := fib.LongestPrefixMatch(0, net.ParseIP("10.50.1.1")); err == nil {
		t.Error("expected route 1 to be withdrawn on BFD DOWN")
	}
	if _, err := fib.LongestPrefixMatch(0, net.ParseIP("10.60.1.1")); err == nil {
		t.Error("expected route 2 to be withdrawn on BFD DOWN")
	}
}

func TestBGPRouteAndPrefixManagement(t *testing.T) {
	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	speaker := NewSpeaker(65001, net.ParseIP("10.0.0.1"), rm)

	neighborIP := net.ParseIP("192.168.40.5")
	cfg := PeerConfig{
		NeighborIP: neighborIP,
		RemoteASN:  65005,
		LocalASN:   65001,
		VRFID:      0,
	}

	_, _ = speaker.AddNeighbor(cfg)

	// Advertise prefix
	_, advNet, _ := net.ParseCIDR("10.244.0.0/16")
	_ = speaker.AdvertisePrefix(neighborIP, 0, *advNet)

	peer, _ := speaker.GetPeer(neighborIP, 0)
	if len(peer.AdvertisedRoutes) != 1 {
		t.Errorf("expected 1 advertised route, got %d", len(peer.AdvertisedRoutes))
	}

	// Receive and then withdraw a specific route
	_, rcvNet, _ := net.ParseCIDR("172.20.0.0/16")
	_ = speaker.ReceiveRoute(neighborIP, 0, *rcvNet, 10)
	if len(peer.ReceivedRoutes) != 1 {
		t.Errorf("expected 1 received route, got %d", len(peer.ReceivedRoutes))
	}

	err := speaker.WithdrawRoute(neighborIP, 0, *rcvNet)
	if err != nil {
		t.Fatalf("withdraw route failed: %v", err)
	}
	if len(peer.ReceivedRoutes) != 0 {
		t.Errorf("expected 0 received routes after withdrawal, got %d", len(peer.ReceivedRoutes))
	}

	// List peers
	peers := speaker.ListPeers()
	if len(peers) != 1 {
		t.Errorf("expected 1 peer in list, got %d", len(peers))
	}
}
