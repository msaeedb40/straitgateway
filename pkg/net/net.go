// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package net provides network utility helpers for straitgateway.
package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
)

// ParseCIDR parses a CIDR string and returns the prefix.
func ParseCIDR(cidr string) (netip.Prefix, error) {
	return netip.ParsePrefix(cidr)
}

// ParseIP parses an IP address string into a netip.Addr.
func ParseIP(ip string) (netip.Addr, error) {
	return netip.ParseAddr(ip)
}

// IPToUint32 converts an IPv4 address to a uint32 (big-endian, for BPF maps).
func IPToUint32(ip netip.Addr) uint32 {
	ip = ip.Unmap()
	if !ip.Is4() {
		return 0
	}
	b := ip.As4()
	return binary.BigEndian.Uint32(b[:])
}

// Uint32ToIP converts a big-endian uint32 to an IPv4 netip.Addr.
func Uint32ToIP(v uint32) netip.Addr {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return netip.AddrFrom4(b)
}

// MACFromBytes converts a 6-byte slice to a net.HardwareAddr.
func MACFromBytes(b [6]byte) net.HardwareAddr {
	return net.HardwareAddr(b[:])
}

// GenerateMAC generates a deterministic link-local MAC address from an index.
func GenerateMAC(index int) [6]byte {
	var mac [6]byte
	// locally administered, unicast prefix
	mac[0] = 0xAA
	mac[1] = 0xBB
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(index))
	copy(mac[2:], b)
	return mac
}

// DiscoverMTU discovers the MTU of the given network interface.
// Returns the interface MTU or a safe default of 1500 on error.
func DiscoverMTU(ifaceName string) int {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return 1500
	}
	return iface.MTU
}

// OverlayMTU calculates the overlay MTU accounting for encapsulation overhead.
// vxlanOverhead = 50, geneveOverhead = 60, gre = 4.
func OverlayMTU(baseMTU int, overhead int) int {
	mtu := baseMTU - overhead
	if mtu < 576 {
		mtu = 576
	}
	return mtu
}

// IsIPv4 returns true if the address is an IPv4 address.
func IsIPv4(addr netip.Addr) bool {
	return addr.Is4() || addr.Is4In6()
}

// IsIPv6 returns true if the address is a native IPv6 address.
func IsIPv6(addr netip.Addr) bool {
	return addr.Is6() && !addr.Is4In6()
}

// ContainsCIDR returns true if outer contains inner.
func ContainsCIDR(outer, inner netip.Prefix) bool {
	return outer.Contains(inner.Addr()) && outer.Bits() <= inner.Bits()
}

// PrefixLen returns the CIDR prefix length as a formatted string.
func PrefixLen(prefix netip.Prefix) string {
	return fmt.Sprintf("%d", prefix.Bits())
}

// NodeIPFromInterface returns the primary IP address of the given interface.
func NodeIPFromInterface(ifaceName string) (netip.Addr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("interface %q not found: %w", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return netip.Addr{}, fmt.Errorf("listing addrs on %q: %w", ifaceName, err)
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip, ok := netip.AddrFromSlice(ipnet.IP)
		if !ok {
			continue
		}
		ip = ip.Unmap()
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		return ip, nil
	}
	return netip.Addr{}, fmt.Errorf("no usable address on interface %q", ifaceName)
}
