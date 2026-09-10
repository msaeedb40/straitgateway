// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
	"github.com/msaeedb40/straitgateway/policy"
)

func TestPolicyEngineDefaultDeny(t *testing.T) {
	engine := policy.NewEngine(nil) // no rules
	action := engine.Evaluate(100, 200, 80, 6)
	if action != ir.PolicyActionDeny {
		t.Errorf("expected default deny, got %d", action)
	}
}

func TestPolicyEngineFirstMatch(t *testing.T) {
	rules := []ir.PolicyIR{
		{SrcIdentity: 100, DstIdentity: 200, DstPort: 80, Protocol: 6, Action: ir.PolicyActionAllow, Priority: 10},
		{SrcIdentity: 100, DstIdentity: 200, DstPort: 80, Protocol: 6, Action: ir.PolicyActionDeny, Priority: 20},
	}
	engine := policy.NewEngine(rules)

	action := engine.Evaluate(100, 200, 80, 6)
	if action != ir.PolicyActionAllow {
		t.Errorf("expected first-match allow, got %d", action)
	}
}

func TestPolicyEngineWildcardMatch(t *testing.T) {
	rules := []ir.PolicyIR{
		{SrcIdentity: 0, DstIdentity: 200, DstPort: 443, Protocol: 6, Action: ir.PolicyActionAllow},
	}
	engine := policy.NewEngine(rules)

	// SrcIdentity=0 means wildcard, should match any source.
	action := engine.Evaluate(999, 200, 443, 6)
	if action != ir.PolicyActionAllow {
		t.Errorf("expected wildcard allow, got %d", action)
	}

	// Different dst should not match → default deny.
	action = engine.Evaluate(999, 300, 443, 6)
	if action != ir.PolicyActionDeny {
		t.Errorf("expected deny for unmatched dst, got %d", action)
	}
}
