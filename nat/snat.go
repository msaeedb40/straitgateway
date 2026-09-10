// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package nat manages SNAT, DNAT, masquerade, and NAT64 rules.
// Rules are compiled to NatIR and applied by the dataplane compiler.
package nat

import (
	"net/netip"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// SNATRule defines a source NAT rule.
type SNATRule struct {
	SrcCIDR  netip.Prefix
	SnatToIP netip.Addr
}

// CompileSNAT compiles SNAT rules to NatIR.
func CompileSNAT(rules []SNATRule) []ir.NatIR {
	var result []ir.NatIR
	for _, r := range rules {
		result = append(result, ir.NatIR{
			SrcCIDR:  r.SrcCIDR,
			SnatToIP: r.SnatToIP,
			Type:     ir.NatTypeSNAT,
		})
	}
	return result
}

// MasqueradeRule defines a masquerade rule for pod traffic leaving the node.
type MasqueradeRule struct {
	SrcCIDR  netip.Prefix
	Protocol sgtypes.Protocol
}

// CompileMasquerade compiles masquerade rules to NatIR.
func CompileMasquerade(rules []MasqueradeRule) []ir.NatIR {
	var result []ir.NatIR
	for _, r := range rules {
		result = append(result, ir.NatIR{
			SrcCIDR:  r.SrcCIDR,
			Protocol: r.Protocol,
			Type:     ir.NatTypeMasquerade,
		})
	}
	return result
}
