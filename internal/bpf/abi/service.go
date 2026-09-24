package abi

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ABI size constants for eBPF map creation (Finding #12)
const (
	ServiceKeySize = 8
	ServiceValSize = 8
	BackendKeySize = 8
	BackendValSize = 8
)

// ServiceKey matches `struct sg_service_key` in bpf/maps/services.bpf.c (size: 8 bytes).
type ServiceKey struct {
	Addr  IPv4Addr // VIP IPv4 address (network byte order)
	Port  uint16   // Service port (network byte order)
	Proto uint8    // IP protocol: ProtoTCP, ProtoUDP
	Pad   uint8    // Alignment padding
}

// NewServiceKey constructs a ServiceKey with ports in network byte order.
func NewServiceKey(vip net.IP, port uint16, proto uint8) (ServiceKey, error) {
	ip4, err := IPv4FromNetIP(vip)
	if err != nil {
		return ServiceKey{}, err
	}
	var netPort [2]byte
	binary.BigEndian.PutUint16(netPort[:], port)

	return ServiceKey{
		Addr:  ip4,
		Port:  binary.BigEndian.Uint16(netPort[:]),
		Proto: proto,
		Pad:   0,
	}, nil
}

// HostPort returns the port in host byte order.
func (k ServiceKey) HostPort() uint16 {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], k.Port)
	return binary.BigEndian.Uint16(b[:])
}

// ServiceVal matches `struct sg_service_val` in bpf/maps/services.bpf.c (size: 8 bytes).
type ServiceVal struct {
	BackendID    uint32 // First backend ID
	BackendCount uint16 // Total number of backends
	Flags        uint16 // ServiceFlag*
}

// BackendKey matches `struct sg_backend_key` in bpf/maps/services.bpf.c (size: 8 bytes).
type BackendKey struct {
	ID   uint32 // Backend / Service ID
	Slot uint32 // Backend slot index (0-based)
}

// BackendVal matches `struct sg_backend_val` in bpf/maps/services.bpf.c (size: 8 bytes).
type BackendVal struct {
	Addr  IPv4Addr // Backend IPv4 address (network byte order)
	Port  uint16   // Backend port (network byte order)
	Flags uint8    // BackendFlag*
	Pad   uint8    // Alignment padding
}

// NewBackendVal creates a BackendVal with network byte order.
func NewBackendVal(endpointIP net.IP, port uint16, flags uint8) (BackendVal, error) {
	ip4, err := IPv4FromNetIP(endpointIP)
	if err != nil {
		return BackendVal{}, err
	}
	var netPort [2]byte
	binary.BigEndian.PutUint16(netPort[:], port)

	return BackendVal{
		Addr:  ip4,
		Port:  binary.BigEndian.Uint16(netPort[:]),
		Flags: flags,
		Pad:   0,
	}, nil
}

// ServiceKeyV6 is the IPv6 service key (size: 20 bytes).
type ServiceKeyV6 struct {
	Addr  IPv6Addr // VIP IPv6 address
	Port  uint16   // Service port
	Proto uint8    // IP protocol
	Pad   uint8    // Alignment
}

// BackendValV6 is the IPv6 backend value (size: 20 bytes).
type BackendValV6 struct {
	Addr  IPv6Addr // Backend IPv6 address
	Port  uint16   // Backend port
	Flags uint8    // Backend state flags
	Pad   uint8    // Alignment
}

func (k ServiceKey) String() string {
	return fmt.Sprintf("%s:%d/%d", k.Addr, k.HostPort(), k.Proto)
}
