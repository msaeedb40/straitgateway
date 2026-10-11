package clustermesh

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

func setupTestMesh(t *testing.T) (*Manager, *wireguard.Manager, *routing.RouteManager) {
	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	wg, err := wireguard.NewManager("wg-test", 51820)
	if err != nil {
		t.Fatalf("failed to create wg manager: %v", err)
	}
	wg.SetRouteManager(rm)

	cfg := DefaultConfig()
	cfg.LocalClusterID = "cluster-local"
	cfg.GlobalCIDR = "240.0.0.0/8"

	mgr, err := NewManager(cfg, wg, rm)
	if err != nil {
		t.Fatalf("failed to create clustermesh manager: %v", err)
	}

	return mgr, wg, rm
}

func TestClusterMesh_AddAndRemovePeer(t *testing.T) {
	mgr, wg, _ := setupTestMesh(t)

	peer := &v1alpha1.ClusterPeering{
		Spec: v1alpha1.ClusterPeeringSpec{
			ClusterID:          "cluster-remote-1",
			WireGuardEndpoint:  "198.51.100.20:51820",
			WireGuardPublicKey: "dummy-key-abc1234567890=",
			PodCIDRs:           []string{"10.244.0.0/16"},
			ServiceCIDRs:       []string{"10.96.0.0/16"},
			VRFID:              0,
		},
	}

	// Add peer
	if err := mgr.AddClusterPeer(peer); err != nil {
		t.Fatalf("AddClusterPeer failed: %v", err)
	}

	// Verify state
	st, found := mgr.GetClusterPeer("cluster-remote-1")
	if !found {
		t.Fatalf("expected peer to be found")
	}
	if st.Phase != "Connected" {
		t.Errorf("expected Connected phase, got %s", st.Phase)
	}
	if len(mgr.ListClusterPeers()) != 1 {
		t.Errorf("expected 1 cluster peer")
	}

	// Verify WireGuard peer
	wgPeer, ok := wg.GetPeer("dummy-key-abc1234567890=")
	if !ok || wgPeer == nil {
		t.Errorf("WireGuard peer not registered")
	}

	// Remove peer
	if err := mgr.RemoveClusterPeer("cluster-remote-1"); err != nil {
		t.Fatalf("RemoveClusterPeer failed: %v", err)
	}

	if _, found := mgr.GetClusterPeer("cluster-remote-1"); found {
		t.Errorf("expected peer to be removed")
	}
	if _, ok := wg.GetPeer("dummy-key-abc1234567890="); ok {
		t.Errorf("expected WireGuard peer to be removed")
	}
}

func TestClusterMesh_ExportAndImportService(t *testing.T) {
	mgr, _, _ := setupTestMesh(t)

	// Export service
	rec, err := mgr.ExportService("payments", "production", []v1alpha1.GlobalServicePort{
		{Name: "http", Protocol: "TCP", Port: 8080},
	}, "ClusterMeshPreferred")
	if err != nil {
		t.Fatalf("ExportService failed: %v", err)
	}

	expectedFQDN := "payments.production.global.straitgateway"
	if rec.GlobalName != expectedFQDN {
		t.Errorf("expected %s, got %s", expectedFQDN, rec.GlobalName)
	}
	if rec.GlobalVIP == nil || !stringsHasPrefix(rec.GlobalVIP.String(), "240.") {
		t.Errorf("expected VIP in 240.0.0.0/8, got %v", rec.GlobalVIP)
	}

	// Resolve VIP
	vip, ok := mgr.ResolveGlobalVIP(expectedFQDN)
	if !ok || !vip.Equal(rec.GlobalVIP) {
		t.Errorf("ResolveGlobalVIP mismatch")
	}

	// Import remote service
	importRec, err := mgr.ImportService(&v1alpha1.GlobalServiceImport{
		Spec: v1alpha1.GlobalServiceImportSpec{
			GlobalServiceName: "inventory.production.global.straitgateway",
			GlobalVIP:         "240.0.10.5",
			Ports: []v1alpha1.GlobalServicePort{
				{Port: 9000},
			},
			Clusters: []v1alpha1.ClusterBackendRef{
				{ClusterID: "cluster-remote-1", EndpointIPs: []string{"10.244.1.15:9000"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("ImportService failed: %v", err)
	}
	if importRec.GlobalVIP.String() != "240.0.10.5" {
		t.Errorf("expected 240.0.10.5 VIP, got %v", importRec.GlobalVIP)
	}
}

func TestClusterMesh_TrafficPolicies(t *testing.T) {
	mgr, _, _ := setupTestMesh(t)

	_, err := mgr.ExportService("orders", "default", []v1alpha1.GlobalServicePort{{Port: 80}}, "ClusterMeshPreferred")
	if err != nil {
		t.Fatalf("ExportService failed: %v", err)
	}

	// Register local and remote endpoints
	_ = mgr.RegisterRemoteEndpoints("orders", "default", "cluster-local", []string{"10.240.0.10:80"})
	_ = mgr.RegisterRemoteEndpoints("orders", "default", "cluster-remote", []string{"10.244.0.20:80"})

	// 1. ClusterMeshPreferred should return local endpoint first
	ep, err := mgr.SelectBackend("orders", "default", "cluster-local")
	if err != nil {
		t.Fatalf("SelectBackend failed: %v", err)
	}
	if ep != "10.240.0.10:80" {
		t.Errorf("expected local endpoint 10.240.0.10:80, got %s", ep)
	}

	// 2. ClusterLocal should only return local
	rec, _ := mgr.GetGlobalService(mgr.FormatGlobalName("orders", "default"))
	rec.TrafficPolicy = "ClusterLocal"
	epLocal, err := mgr.SelectBackend("orders", "default", "cluster-local")
	if err != nil || epLocal != "10.240.0.10:80" {
		t.Errorf("expected local endpoint, got %s (err: %v)", epLocal, err)
	}

	// ClusterLocal from a cluster with no endpoints should fail
	_, err = mgr.SelectBackend("orders", "default", "cluster-nonexistent")
	if err == nil {
		t.Errorf("expected error for cluster with no local endpoints")
	}

	// 3. Fallback when local endpoints empty in ClusterMeshPreferred
	rec.TrafficPolicy = "ClusterMeshPreferred"
	rec.ClusterBackends["cluster-local"] = nil
	epFallback, err := mgr.SelectBackend("orders", "default", "cluster-local")
	if err != nil || epFallback != "10.244.0.20:80" {
		t.Errorf("expected fallback to remote endpoint 10.244.0.20:80, got %s", epFallback)
	}
}

func TestClusterMesh_DNSInterceptor(t *testing.T) {
	mgr, _, _ := setupTestMesh(t)

	rec, _ := mgr.ExportService("cart", "store", []v1alpha1.GlobalServicePort{{Port: 80}}, "RoundRobin")
	interceptor := NewDNSInterceptor(mgr, 5353)

	// Build DNS query packet for "cart.store.global.straitgateway"
	queryPacket := buildMockDNSQuery(0x1234, "cart.store.global.straitgateway", 1 /* A */)

	resp, err := interceptor.ProcessDNSQuery(queryPacket)
	if err != nil {
		t.Fatalf("ProcessDNSQuery failed: %v", err)
	}

	if len(resp) < 16 {
		t.Fatalf("response too short: %d", len(resp))
	}

	// Check response txID
	txID := binary.BigEndian.Uint16(resp[0:2])
	if txID != 0x1234 {
		t.Errorf("expected txID 0x1234, got 0x%x", txID)
	}

	// Check flags: QR=1 (bit 15), RCODE=0
	flags := binary.BigEndian.Uint16(resp[2:4])
	if flags&0x8000 == 0 {
		t.Errorf("expected QR=1 (response flag)")
	}
	if flags&0x000F != 0 {
		t.Errorf("expected RCODE=0 (NoError), got %d", flags&0x000F)
	}

	// Verify synthesized IP matches VIP
	ipOffset := len(resp) - 4
	respIP := net.IP(resp[ipOffset:])
	if !respIP.Equal(rec.GlobalVIP) {
		t.Errorf("expected DNS answer %v, got %v", rec.GlobalVIP, respIP)
	}

	// Test NXDOMAIN for unregistered service
	nxQuery := buildMockDNSQuery(0x5678, "unknown.store.global.straitgateway", 1)
	nxResp, err := interceptor.ProcessDNSQuery(nxQuery)
	if err != nil {
		t.Fatalf("NXDOMAIN query failed: %v", err)
	}
	nxFlags := binary.BigEndian.Uint16(nxResp[2:4])
	if nxFlags&0x000F != 3 { // RCODE 3 is NXDOMAIN
		t.Errorf("expected NXDOMAIN (rcode 3), got %d", nxFlags&0x000F)
	}
}

func buildMockDNSQuery(txID uint16, fqdn string, qtype uint16) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, txID)
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x0100)) // Standard query, RD=1
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))      // QDCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))      // ANCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))      // NSCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))      // ARCount

	buf.Write(encodeDomainName(fqdn))
	_ = binary.Write(&buf, binary.BigEndian, qtype)
	_ = binary.Write(&buf, binary.BigEndian, uint16(1)) // QCLASS: IN
	return buf.Bytes()
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
