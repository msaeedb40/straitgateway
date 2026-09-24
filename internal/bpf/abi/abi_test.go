package abi

import (
	"net"
	"testing"
	"unsafe"
)

func TestStructSizesAndAlignment(t *testing.T) {
	tests := []struct {
		name     string
		gotSize  uintptr
		wantSize uintptr
	}{
		{"IPv4Addr", unsafe.Sizeof(IPv4Addr{}), 4},
		{"IPv6Addr", unsafe.Sizeof(IPv6Addr{}), 16},
		{"IPAddr", unsafe.Sizeof(IPAddr{}), 20},
		{"VersionHeader", unsafe.Sizeof(VersionHeader{}), 16},
		{"ServiceKey", unsafe.Sizeof(ServiceKey{}), 8},
		{"ServiceVal", unsafe.Sizeof(ServiceVal{}), 8},
		{"ServiceKeyV6", unsafe.Sizeof(ServiceKeyV6{}), 20},
		{"BackendKey", unsafe.Sizeof(BackendKey{}), 8},
		{"BackendVal", unsafe.Sizeof(BackendVal{}), 8},
		{"BackendValV6", unsafe.Sizeof(BackendValV6{}), 20},
		{"RouteKey", unsafe.Sizeof(RouteKey{}), 8},
		{"LPMRouteKey", unsafe.Sizeof(LPMRouteKey{}), 8},
		{"RouteVal", unsafe.Sizeof(RouteVal{}), 12},
		{"RouteKeyV6", unsafe.Sizeof(RouteKeyV6{}), 20},
		{"RouteValV6", unsafe.Sizeof(RouteValV6{}), 24},
		{"PolicyKey", unsafe.Sizeof(PolicyKey{}), 12},
		{"PolicyVal", unsafe.Sizeof(PolicyVal{}), 4},
		{"PeerKey", unsafe.Sizeof(PeerKey{}), 4},
		{"PeerVal", unsafe.Sizeof(PeerVal{}), 8},
		{"PeerValV6", unsafe.Sizeof(PeerValV6{}), 20},
		{"FlowKey", unsafe.Sizeof(FlowKey{}), 16},
		{"FlowKeyV6", unsafe.Sizeof(FlowKeyV6{}), 40},
		{"FlowVal", unsafe.Sizeof(FlowVal{}), 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotSize != tt.wantSize {
				t.Errorf("struct %s size mismatch: got %d bytes, want %d bytes",
					tt.name, tt.gotSize, tt.wantSize)
			}
		})
	}
}

func TestIPConversions(t *testing.T) {
	// IPv4
	ipStr := "10.244.1.42"
	v4, err := ParseIPv4(ipStr)
	if err != nil {
		t.Fatalf("ParseIPv4 failed: %v", err)
	}
	if v4.String() != ipStr {
		t.Errorf("got %s, want %s", v4.String(), ipStr)
	}
	if v4.AsUint32() == 0 {
		t.Errorf("AsUint32 should not be 0")
	}

	// IPv6
	v6Str := "fd00:10:244::1"
	v6, err := ParseIPv6(v6Str)
	if err != nil {
		t.Fatalf("ParseIPv6 failed: %v", err)
	}
	if v6.AsNetIP().String() != "fd00:10:244::1" {
		t.Errorf("got %s, want fd00:10:244::1", v6.String())
	}

	// Generic IPAddr
	unified4, err := FromNetIP(net.ParseIP(ipStr))
	if err != nil {
		t.Fatalf("FromNetIP failed: %v", err)
	}
	if unified4.Family != FamilyIPv4 {
		t.Errorf("got family %d, want FamilyIPv4 (%d)", unified4.Family, FamilyIPv4)
	}
	if unified4.String() != ipStr {
		t.Errorf("got %s, want %s", unified4.String(), ipStr)
	}

	unified6, err := FromNetIP(net.ParseIP(v6Str))
	if err != nil {
		t.Fatalf("FromNetIP v6 failed: %v", err)
	}
	if unified6.Family != FamilyIPv6 {
		t.Errorf("got family %d, want FamilyIPv6 (%d)", unified6.Family, FamilyIPv6)
	}
}

func TestServiceABIConstructors(t *testing.T) {
	vip := net.ParseIP("10.96.0.10")
	svcKey, err := NewServiceKey(vip, 53, ProtoUDP)
	if err != nil {
		t.Fatalf("NewServiceKey failed: %v", err)
	}
	if svcKey.HostPort() != 53 {
		t.Errorf("HostPort: got %d, want 53", svcKey.HostPort())
	}
	if svcKey.Proto != ProtoUDP {
		t.Errorf("Proto: got %d, want %d", svcKey.Proto, ProtoUDP)
	}

	epIP := net.ParseIP("10.244.0.5")
	bVal, err := NewBackendVal(epIP, 53, BackendFlagActive)
	if err != nil {
		t.Fatalf("NewBackendVal failed: %v", err)
	}
	if bVal.Flags != BackendFlagActive {
		t.Errorf("Flags: got %d, want %d", bVal.Flags, BackendFlagActive)
	}
}

func TestPolicyABI(t *testing.T) {
	pk := NewPolicyKey(100, 200, 8080, ProtoTCP, DirIngress)
	if pk.SrcID != 100 || pk.DstID != 200 {
		t.Errorf("Src/Dst IDs unexpected: %+v", pk)
	}
	if pk.HostPort() != 8080 {
		t.Errorf("HostPort: got %d, want 8080", pk.HostPort())
	}
	if pk.Dir != DirIngress {
		t.Errorf("Dir: got %d, want %d", pk.Dir, DirIngress)
	}
}

func TestPeerAndFlowABI(t *testing.T) {
	peerKey := PeerKey{PeerID: 42}
	if peerKey.PeerID != 42 {
		t.Errorf("expected PeerID 42, got %d", peerKey.PeerID)
	}

	flowKey := FlowKey{
		SrcPort: 12345,
		DstPort: 80,
		Proto:   ProtoTCP,
	}
	if flowKey.DstPort != 80 {
		t.Errorf("expected DstPort 80, got %d", flowKey.DstPort)
	}

	flowVal := FlowVal{
		Packets:    1000,
		Bytes:      500000,
		LastSeenNs: 123456789,
		SrcID:      10,
		DstID:      20,
		State:      1,
	}
	if flowVal.Packets != 1000 || flowVal.Bytes != 500000 {
		t.Errorf("unexpected flowVal counters: %+v", flowVal)
	}
}

func TestABIVersionValidation(t *testing.T) {
	valid := VersionHeader{
		Magic:      ABIMagic,
		Version:    CurrentABIVersion,
		Generation: 1,
	}
	if err := ValidateVersion(valid); err != nil {
		t.Errorf("ValidateVersion failed on valid header: %v", err)
	}

	invalidMagic := valid
	invalidMagic.Magic = 0x1234
	if err := ValidateVersion(invalidMagic); err == nil {
		t.Errorf("ValidateVersion should fail on invalid magic")
	}

	invalidVer := valid
	invalidVer.Version = 99
	if err := ValidateVersion(invalidVer); err == nil {
		t.Errorf("ValidateVersion should fail on invalid version")
	}
}
