package wireguard

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/internal/routing"
)

func TestWireGuardManager(t *testing.T) {
	mgr, err := NewManager("wg0", 51820)
	if err != nil {
		t.Fatalf("failed to create wireguard manager: %v", err)
	}

	pubKey := mgr.GetPublicKey()
	if pubKey == "" {
		t.Fatal("expected non-empty public key")
	}

	_, allowedNet, _ := net.ParseCIDR("10.200.0.0/16")
	peer := PeerConfig{
		PublicKey:           "test-remote-public-key",
		Endpoint:            "192.168.1.100:51820",
		AllowedIPs:         []net.IPNet{*allowedNet},
		PersistentKeepalive: 25,
	}

	if err := mgr.AddPeer(peer); err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	peers := mgr.GetPeers()
	if len(peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(peers))
	}

	if err := mgr.RemovePeer(peer.PublicKey); err != nil {
		t.Fatalf("failed to remove peer: %v", err)
	}

	if len(mgr.GetPeers()) != 0 {
		t.Errorf("expected 0 peers after removal, got %d", len(mgr.GetPeers()))
	}
}

func TestWireGuardRouteIntegration(t *testing.T) {
	mgr, err := NewManager("wg0", 51820)
	if err != nil {
		t.Fatalf("failed to create wireguard manager: %v", err)
	}

	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	mgr.SetRouteManager(rm)

	_, allowedNet, _ := net.ParseCIDR("10.210.0.0/16")
	peer := PeerConfig{
		PublicKey:   "cluster-b-pubkey",
		AllowedIPs:  []net.IPNet{*allowedNet},
	}

	// Add peer and check route automatically created
	if err := mgr.AddPeer(peer); err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	matched, err := fib.LongestPrefixMatch(0, net.ParseIP("10.210.15.1"))
	if err != nil {
		t.Fatalf("expected route to be automatically installed for WireGuard peer: %v", err)
	}
	if matched.Interface != "wg0" {
		t.Errorf("expected interface wg0, got %s", matched.Interface)
	}

	// GetPeer test
	p, found := mgr.GetPeer("cluster-b-pubkey")
	if !found || p == nil {
		t.Error("expected to find peer by public key")
	}

	// Remove peer and check route automatically withdrawn
	if err := mgr.RemovePeer(peer.PublicKey); err != nil {
		t.Fatalf("failed to remove peer: %v", err)
	}

	_, err = fib.LongestPrefixMatch(0, net.ParseIP("10.210.15.1"))
	if err == nil {
		t.Error("expected route to be withdrawn after WireGuard peer removal")
	}
}

