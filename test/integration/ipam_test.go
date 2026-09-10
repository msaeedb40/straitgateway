// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/netip"
	"testing"

	"github.com/msaeedb40/straitgateway/ipam"
)

func TestIPAMAllocate(t *testing.T) {
	cidr := netip.MustParsePrefix("10.244.1.0/24")
	alloc, err := ipam.New(cidr)
	if err != nil {
		t.Fatalf("ipam.New: %v", err)
	}

	// First allocation should be .2 (after network .0 and gateway .1).
	ip, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	expected := netip.MustParseAddr("10.244.1.2")
	if ip != expected {
		t.Errorf("expected %s, got %s", expected, ip)
	}

	// Gateway should be .1.
	gw := alloc.Gateway()
	expectedGW := netip.MustParseAddr("10.244.1.1")
	if gw != expectedGW {
		t.Errorf("gateway: expected %s, got %s", expectedGW, gw)
	}
}

func TestIPAMExhaustion(t *testing.T) {
	cidr := netip.MustParsePrefix("10.244.1.0/30") // Only 4 addresses total.
	alloc, err := ipam.New(cidr)
	if err != nil {
		t.Fatalf("ipam.New: %v", err)
	}

	// /30 = 4 addresses: .0 (net), .1 (gw), .2 (allocatable), .3 (broadcast)
	ip, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("first Allocate: %v", err)
	}
	t.Logf("allocated: %s", ip)

	// Next allocation should fail.
	_, err = alloc.Allocate()
	if err == nil {
		t.Error("expected exhaustion error, got nil")
	}
}

func TestIPAMRelease(t *testing.T) {
	cidr := netip.MustParsePrefix("10.244.1.0/24")
	alloc, err := ipam.New(cidr)
	if err != nil {
		t.Fatalf("ipam.New: %v", err)
	}

	ip1, _ := alloc.Allocate()
	ip2, _ := alloc.Allocate()
	t.Logf("allocated: %s, %s", ip1, ip2)

	avail := alloc.Available()
	alloc.Release(ip1)
	if alloc.Available() != avail+1 {
		t.Error("release did not increase available count")
	}
}
