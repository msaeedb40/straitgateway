package wireguard

import (
	"net"
	"testing"
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
