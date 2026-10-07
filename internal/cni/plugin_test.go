package cni

import (
	"net"
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
