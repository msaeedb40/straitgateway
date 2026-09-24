// Package abi defines the explicit, versioned Go ↔ eBPF memory layouts
// and serialization contracts for StraitGateway eBPF maps and programs.
//
// These structs must strictly match the C struct definitions in bpf/maps/
// and bpf/headers/ to prevent size, padding, or endianness mismatches between
// userspace Go controllers and kernel eBPF datapath.
package abi

import (
	"fmt"
)

// CurrentABIVersion represents the current BPF ABI contract version.
const CurrentABIVersion uint32 = 1

// Magic identifier for StraitGateway BPF ABI validation.
const ABIMagic uint32 = 0x53474159 // "SGAY"

// Protocol numbers matching Linux socket / IP protocols.
const (
	ProtoTCP  uint8 = 6
	ProtoUDP  uint8 = 17
	ProtoSCTP uint8 = 132
	ProtoICMP uint8 = 1
	ProtoICMP6 uint8 = 58
)

// Direction constants for policy and flow maps.
const (
	DirIngress uint8 = 0
	DirEgress  uint8 = 1
)

// Policy verdicts.
const (
	VerdictDeny  uint8 = 0
	VerdictAllow uint8 = 1
)

// Service flags matching sg_service_val.flags.
const (
	ServiceFlagActive        uint16 = 1 << 0
	ServiceFlagSessionAff    uint16 = 1 << 1
	ServiceFlagNodePort      uint16 = 1 << 2
	ServiceFlagExternalIP    uint16 = 1 << 3
	ServiceFlagLoadBalancer  uint16 = 1 << 4
)

// Backend flags matching sg_backend_val.flags.
const (
	BackendFlagActive uint8 = 1 << 0
	BackendFlagDrain  uint8 = 1 << 1
)

// Route flags matching sg_route_val.flags.
const (
	RouteFlagBlackhole uint8 = 1 << 0
	RouteFlagLocal     uint8 = 1 << 1
)

// VersionHeader can be embedded in shared maps to verify userspace/kernel synchronization.
type VersionHeader struct {
	Magic      uint32
	Version    uint32
	Generation uint64
}

// ValidateVersion checks whether the given header matches the supported ABI.
func ValidateVersion(hdr VersionHeader) error {
	if hdr.Magic != ABIMagic {
		return fmt.Errorf("invalid ABI magic: expected 0x%X, got 0x%X", ABIMagic, hdr.Magic)
	}
	if hdr.Version != CurrentABIVersion {
		return fmt.Errorf("incompatible ABI version: expected %d, got %d", CurrentABIVersion, hdr.Version)
	}
	return nil
}
