package observability

import (
	"net"
	"strings"
	"testing"

	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

func TestMetricsCollector(t *testing.T) {
	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	bgpSpeaker := bgp.NewSpeaker(65001, net.ParseIP("10.0.0.1"), rm)
	bfdMon := bfd.NewMonitor()
	defer bfdMon.Close()

	wgMgr, _ := wireguard.NewManager("wg0", 51820)

	// Add an established BGP neighbor
	_, _ = bgpSpeaker.AddNeighbor(bgp.PeerConfig{
		NeighborIP: net.ParseIP("192.168.1.2"),
		LocalASN:   65001,
		RemoteASN:  65002,
	})

	// Add an active BFD session
	_ = bfdMon.AddSession(bfd.SessionConfig{
		PeerIP: net.ParseIP("192.168.1.2"),
	}, nil)

	// Add a WireGuard peer
	_, allowed, _ := net.ParseCIDR("10.100.0.0/16")
	_ = wgMgr.AddPeer(wireguard.PeerConfig{
		PublicKey:  "mock-peer-key",
		AllowedIPs: []net.IPNet{*allowed},
	})

	col := NewMetricsCollector(nil, bgpSpeaker, bfdMon, wgMgr, "node-1")
	m := col.Collect()

	if m.BGPSessionsActive != 1 {
		t.Errorf("expected 1 active BGP session, got %d", m.BGPSessionsActive)
	}
	if m.BFDSessionsActive != 1 {
		t.Errorf("expected 1 active BFD session, got %d", m.BFDSessionsActive)
	}
	if m.WireGuardPeersCount != 1 {
		t.Errorf("expected 1 WireGuard peer, got %d", m.WireGuardPeersCount)
	}

	// Verify Prometheus text export
	promText := col.PrometheusExport()
	if !strings.Contains(promText, "straitgateway_packets_total{node=\"node-1\"}") {
		t.Error("expected straitgateway_packets_total metric in output")
	}
	if !strings.Contains(promText, "straitgateway_bgp_sessions{node=\"node-1\"} 1") {
		t.Error("expected straitgateway_bgp_sessions with value 1 in output")
	}
	if !strings.Contains(promText, "straitgateway_wireguard_peers{node=\"node-1\"} 1") {
		t.Error("expected straitgateway_wireguard_peers with value 1 in output")
	}
}
