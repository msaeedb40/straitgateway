package cni

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestParseNetConf(t *testing.T) {
	raw := []byte(`{
		"cniVersion": "1.0.0",
		"name": "straitgateway-net",
		"type": "sg-cni",
		"subnetV4": "10.244.2.0/24",
		"mtu": 1450
	}`)

	conf, err := ParseNetConf(raw)
	if err != nil {
		t.Fatalf("failed to parse netconf: %v", err)
	}

	if conf.CNIVersion != "1.0.0" {
		t.Errorf("expected cniVersion 1.0.0, got %s", conf.CNIVersion)
	}
	if conf.SubnetV4 != "10.244.2.0/24" {
		t.Errorf("expected subnet 10.244.2.0/24, got %s", conf.SubnetV4)
	}
	if conf.MTU != 1450 {
		t.Errorf("expected MTU 1450, got %d", conf.MTU)
	}
}

func TestDeriveGateway(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("10.244.5.0/24")
	if err != nil {
		t.Fatalf("parse CIDR error: %v", err)
	}

	gw := deriveGateway(*ipNet)
	if gw == nil {
		t.Fatal("expected non-nil gateway")
	}
	if gw.String() != "10.244.5.1" {
		t.Errorf("expected gateway 10.244.5.1, got %s", gw.String())
	}
}

func TestDeriveHostIfName(t *testing.T) {
	cID := "c1234567890abcdef"
	name1 := DeriveHostIfName(cID)
	name2 := DeriveHostIfName(cID)

	if name1 != name2 {
		t.Errorf("expected deterministic host ifname, got %s and %s", name1, name2)
	}
	if len(name1) > 15 {
		t.Errorf("host ifname too long for IFNAMSIZ: %s (len %d)", name1, len(name1))
	}
	if name1[:2] != "sg" {
		t.Errorf("expected sg prefix, got %s", name1)
	}
}

func TestDerivePeerIfName(t *testing.T) {
	cID := "c1234567890abcdef"
	name1 := DerivePeerIfName(cID)
	name2 := DerivePeerIfName(cID)

	if name1 != name2 {
		t.Errorf("expected deterministic peer ifname, got %s and %s", name1, name2)
	}
	if len(name1) > 15 {
		t.Errorf("peer ifname too long for IFNAMSIZ: %s (len %d)", name1, len(name1))
	}
	if name1[:3] != "sgp" {
		t.Errorf("expected sgp prefix, got %s", name1)
	}
	// Verify that host interface name and peer interface name are strictly distinct
	hostName := DeriveHostIfName(cID)
	if name1 == hostName {
		t.Errorf("peer ifname and host ifname must never be equal: %s == %s", name1, hostName)
	}
	if name1 == "eth0" || hostName == "eth0" {
		t.Errorf("derived ifnames must never collide with host default eth0")
	}
}

func TestGenerateAndInstallCNIConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cni-conf-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data, err := GenerateConflist("10.244.0.0/16", "fd00:10:244::/64", 1450)
	if err != nil {
		t.Fatalf("generate conflist failed: %v", err)
	}

	var conflist CNIConflist
	if err := json.Unmarshal(data, &conflist); err != nil {
		t.Fatalf("invalid generated conflist json: %v", err)
	}
	if conflist.Name != "straitgateway" {
		t.Errorf("expected conflist name straitgateway, got %s", conflist.Name)
	}
	if len(conflist.Plugins) != 2 {
		t.Errorf("expected 2 plugins in conflist, got %d", len(conflist.Plugins))
	}

	err = InstallCNIConfig(tmpDir, "10.244.0.0/16", "", 1500)
	if err != nil {
		t.Fatalf("install CNI config failed: %v", err)
	}

	installedPath := filepath.Join(tmpDir, DefaultCNIConfigName)
	if _, err := os.Stat(installedPath); os.IsNotExist(err) {
		t.Fatalf("installed conflist file missing: %s", installedPath)
	}
}
