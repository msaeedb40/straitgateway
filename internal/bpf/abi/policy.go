package abi

import (
	"encoding/binary"
	"fmt"
)

// ABI size constants for eBPF map creation (Finding #12)
const (
	PolicyKeySize   = 12
	PolicyValSize   = 4
	IdentityKeySize = 4
	IdentityValSize = 4
)

// PolicyKey matches `struct sg_policy_key` in bpf/maps/policy.bpf.c (size: 12 bytes).
type PolicyKey struct {
	SrcID   uint32 // Source identity (eBPF-assigned security ID)
	DstID   uint32 // Destination identity
	DstPort uint16 // Destination L4 port (network byte order)
	Proto   uint8  // L4 Protocol: ProtoTCP, ProtoUDP, etc.
	Dir     uint8  // DirIngress (0) or DirEgress (1)
}

// NewPolicyKey creates a PolicyKey with port encoded in network byte order.
func NewPolicyKey(srcID, dstID uint32, port uint16, proto, dir uint8) PolicyKey {
	var netPort [2]byte
	binary.BigEndian.PutUint16(netPort[:], port)

	return PolicyKey{
		SrcID:   srcID,
		DstID:   dstID,
		DstPort: binary.BigEndian.Uint16(netPort[:]),
		Proto:   proto,
		Dir:     dir,
	}
}

// HostPort returns the destination port in host byte order.
func (k PolicyKey) HostPort() uint16 {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], k.DstPort)
	return binary.BigEndian.Uint16(b[:])
}

// PolicyVal matches `struct sg_policy_val` in bpf/maps/policy.bpf.c (size: 4 bytes).
type PolicyVal struct {
	Verdict uint8   // VerdictAllow (1) or VerdictDeny (0)
	Pad     [3]byte // Alignment padding to 4 bytes
}

// IdentityKey represents the key for sg_identity_map (netkit ifindex).
type IdentityKey = uint32

// IdentityVal represents the value for sg_identity_map (security identity ID).
type IdentityVal = uint32

func (k PolicyKey) String() string {
	dirStr := "ingress"
	if k.Dir == DirEgress {
		dirStr = "egress"
	}
	return fmt.Sprintf("policy[%s src=%d dst=%d port=%d proto=%d]", dirStr, k.SrcID, k.DstID, k.HostPort(), k.Proto)
}
