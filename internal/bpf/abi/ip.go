package abi

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
)

// Address family constants matching Linux AF_INET and AF_INET6.
const (
	FamilyIPv4 uint8 = 2  // syscall.AF_INET
	FamilyIPv6 uint8 = 10 // syscall.AF_INET6
)

// IPv4Addr is a fixed 4-byte IPv4 address in network byte order.
type IPv4Addr [4]byte

// IPv4FromNetIP creates an IPv4Addr from a net.IP.
func IPv4FromNetIP(ip net.IP) (IPv4Addr, error) {
	ip4 := ip.To4()
	if ip4 == nil {
		return IPv4Addr{}, fmt.Errorf("not an IPv4 address: %v", ip)
	}
	var a IPv4Addr
	copy(a[:], ip4)
	return a, nil
}

// IPv4FromUint32 creates an IPv4Addr from a big-endian uint32.
func IPv4FromUint32(v uint32) IPv4Addr {
	var a IPv4Addr
	binary.BigEndian.PutUint32(a[:], v)
	return a
}

// ParseIPv4 parses an IPv4 string.
func ParseIPv4(s string) (IPv4Addr, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return IPv4Addr{}, fmt.Errorf("invalid IP string: %q", s)
	}
	return IPv4FromNetIP(ip)
}

// AsUint32 returns the IPv4Addr as a big-endian uint32.
func (a IPv4Addr) AsUint32() uint32 {
	return binary.BigEndian.Uint32(a[:])
}

// AsNetIP returns a net.IP representation.
func (a IPv4Addr) AsNetIP() net.IP {
	ip := make(net.IP, 4)
	copy(ip, a[:])
	return ip
}

// String returns the dot-decimal representation.
func (a IPv4Addr) String() string {
	return a.AsNetIP().String()
}

// IPv6Addr is a fixed 16-byte IPv6 address in network byte order.
type IPv6Addr [16]byte

// IPv6FromNetIP creates an IPv6Addr from a net.IP.
func IPv6FromNetIP(ip net.IP) (IPv6Addr, error) {
	ip16 := ip.To16()
	if ip16 == nil {
		return IPv6Addr{}, fmt.Errorf("not an IPv6 address: %v", ip)
	}
	var a IPv6Addr
	copy(a[:], ip16)
	return a, nil
}

// ParseIPv6 parses an IPv6 string.
func ParseIPv6(s string) (IPv6Addr, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return IPv6Addr{}, fmt.Errorf("invalid IP string: %q", s)
	}
	return IPv6FromNetIP(ip)
}

// AsNetIP returns a net.IP representation.
func (a IPv6Addr) AsNetIP() net.IP {
	ip := make(net.IP, 16)
	copy(ip, a[:])
	return ip
}

// String returns the standard IPv6 colon-hex string representation.
func (a IPv6Addr) String() string {
	return a.AsNetIP().String()
}

// IPAddr is a 16-byte fixed-size union-like struct for eBPF maps that support dual stack.
// If Family == FamilyIPv4, only the first 4 bytes of Raw are used.
// If Family == FamilyIPv6, all 16 bytes of Raw are used.
type IPAddr struct {
	Family uint8
	Pad    [3]byte
	Raw    [16]byte
}

// NewIPAddr creates an IPAddr from netip.Addr.
func NewIPAddr(addr netip.Addr) IPAddr {
	var res IPAddr
	if addr.Is4() {
		res.Family = FamilyIPv4
		b := addr.As4()
		copy(res.Raw[:], b[:])
	} else if addr.Is6() {
		res.Family = FamilyIPv6
		b := addr.As16()
		copy(res.Raw[:], b[:])
	}
	return res
}

// FromNetIP converts net.IP to IPAddr.
func FromNetIP(ip net.IP) (IPAddr, error) {
	if ip4 := ip.To4(); ip4 != nil {
		var res IPAddr
		res.Family = FamilyIPv4
		copy(res.Raw[:], ip4)
		return res, nil
	}
	if ip16 := ip.To16(); ip16 != nil {
		var res IPAddr
		res.Family = FamilyIPv6
		copy(res.Raw[:], ip16)
		return res, nil
	}
	return IPAddr{}, fmt.Errorf("unrecognized IP format: %v", ip)
}

// AsNetIP converts IPAddr to net.IP.
func (a IPAddr) AsNetIP() net.IP {
	if a.Family == FamilyIPv4 {
		ip := make(net.IP, 4)
		copy(ip, a.Raw[:4])
		return ip
	}
	ip := make(net.IP, 16)
	copy(ip, a.Raw[:])
	return ip
}

// String returns the string format of IPAddr.
func (a IPAddr) String() string {
	return a.AsNetIP().String()
}
