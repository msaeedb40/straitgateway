// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package nat

import (
	"net/netip"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// DNATRule defines a destination NAT rule.
type DNATRule struct {
	DstIP     netip.Addr
	DstPort   uint16
	NatToIP   netip.Addr
	NatToPort uint16
	Protocol  sgtypes.Protocol
}

// CompileDNAT compiles DNAT rules to NatIR.
func CompileDNAT(rules []DNATRule) []ir.NatIR {
	var result []ir.NatIR
	for _, r := range rules {
		result = append(result, ir.NatIR{
			DstIP:     r.DstIP,
			DstPort:   r.DstPort,
			NatToIP:   r.NatToIP,
			NatToPort: r.NatToPort,
			Protocol:  r.Protocol,
			Type:      ir.NatTypeDNAT,
		})
	}
	return result
}
