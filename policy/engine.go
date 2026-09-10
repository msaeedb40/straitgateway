// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/msaeedb40/straitgateway/dataplane/ir"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// Engine evaluates compiled policy rules against a packet context.
// This is used for debugging/inspection only — actual enforcement is in eBPF.
//
// Rules are evaluated in priority order (0 = highest priority).
// First match wins. Default action is Deny (zero-trust).
type Engine struct {
	rules []ir.PolicyIR
}

// NewEngine creates a new policy evaluation engine from compiled rules.
func NewEngine(rules []ir.PolicyIR) *Engine {
	return &Engine{rules: rules}
}

// Evaluate returns the action for a given packet context.
// Returns PolicyActionDeny if no rule matches (default deny / zero-trust).
func (e *Engine) Evaluate(srcIdentity, dstIdentity sgtypes.Identity, dstPort uint16, protocol uint8) ir.PolicyAction {
	for _, rule := range e.rules {
		if rule.SrcIdentity != 0 && rule.SrcIdentity != srcIdentity {
			continue
		}
		if rule.DstIdentity != 0 && rule.DstIdentity != dstIdentity {
			continue
		}
		if rule.DstPort != 0 && rule.DstPort != dstPort {
			continue
		}
		if rule.Protocol != 0 && uint8(rule.Protocol) != protocol {
			continue
		}

		// First match.
		return rule.Action
	}

	// Default deny.
	return ir.PolicyActionDeny
}
