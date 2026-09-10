// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package policy compiles StraitNetworkPolicy and Kubernetes NetworkPolicy
// resources into PolicyIR objects for the dataplane compiler.
//
// Architectural invariants:
//   - Policy rules are first-match, priority 0–255 (lower = higher priority).
//   - Default action is Deny (zero-trust). No-policy = deny all.
//   - PolicyIR is produced here; BPF map writes happen ONLY in the compiler.
package policy

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"go.uber.org/zap"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	"github.com/msaeedb40/straitgateway/dataplane/ir"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// Compiler compiles policy CRDs into PolicyIR.
type Compiler struct {
	client client.Client
	log    *zap.Logger
}

// New creates a new policy Compiler.
func New(c client.Client, log *zap.Logger) *Compiler {
	return &Compiler{client: c, log: log}
}

// Compile reads all StraitNetworkPolicy and NetworkPolicy resources and
// returns the full ordered set of PolicyIR rules.
func (c *Compiler) Compile(ctx context.Context) ([]ir.PolicyIR, error) {
	var policies []ir.PolicyIR

	// Compile StraitNetworkPolicy resources (higher priority, richer selectors).
	var snpList sgv1.StraitNetworkPolicyList
	if err := c.client.List(ctx, &snpList); err != nil {
		return nil, fmt.Errorf("listing StraitNetworkPolicies: %w", err)
	}
	for _, snp := range snpList.Items {
		compiled, err := c.compileStraitNetworkPolicy(snp)
		if err != nil {
			c.log.Warn("failed to compile StraitNetworkPolicy",
				zap.String("name", snp.Name),
				zap.String("namespace", snp.Namespace),
				zap.Error(err),
			)
			continue
		}
		policies = append(policies, compiled...)
	}

	// Compile standard Kubernetes NetworkPolicy resources.
	var npList networkingv1.NetworkPolicyList
	if err := c.client.List(ctx, &npList); err != nil {
		return nil, fmt.Errorf("listing NetworkPolicies: %w", err)
	}
	for _, np := range npList.Items {
		compiled := c.compileNetworkPolicy(np)
		policies = append(policies, compiled...)
	}

	c.log.Info("policy compilation complete", zap.Int("rules", len(policies)))
	return policies, nil
}

// compileStraitNetworkPolicy compiles a StraitNetworkPolicy into PolicyIR rules.
func (c *Compiler) compileStraitNetworkPolicy(snp sgv1.StraitNetworkPolicy) ([]ir.PolicyIR, error) {
	var rules []ir.PolicyIR
	var policyID sgtypes.PolicyID = 1

	for _, rule := range snp.Spec.Rules {
		pol := ir.PolicyIR{
			ID:       policyID,
			Priority: snp.Spec.Priority,
		}

		// Map action.
		switch rule.Action {
		case sgv1.PolicyActionAllow:
			pol.Action = ir.PolicyActionAllow
		case sgv1.PolicyActionDeny:
			pol.Action = ir.PolicyActionDeny
		case sgv1.PolicyActionReject:
			pol.Action = ir.PolicyActionReject
		}

		// Map direction.
		switch rule.Direction {
		case sgv1.PolicyDirectionIngress:
			pol.Direction = ir.PolicyDirectionIngress
		case sgv1.PolicyDirectionEgress:
			pol.Direction = ir.PolicyDirectionEgress
		}

		// Map ports.
		if len(rule.Ports) > 0 {
			p := rule.Ports[0]
			if p.Port != nil {
				pol.DstPort = uint16(*p.Port)
			}
			switch p.Protocol {
			case sgv1.PolicyProtocolUDP:
				pol.Protocol = sgtypes.ProtocolUDP
			case sgv1.PolicyProtocolICMP:
				pol.Protocol = sgtypes.ProtocolICMP
			default:
				pol.Protocol = sgtypes.ProtocolTCP
			}
		}

		rules = append(rules, pol)
		policyID++
	}

	return rules, nil
}

// compileNetworkPolicy compiles a standard Kubernetes NetworkPolicy into PolicyIR.
func (c *Compiler) compileNetworkPolicy(np networkingv1.NetworkPolicy) []ir.PolicyIR {
	// Standard NetworkPolicy → Allow rules. Default is Deny (enforced by eBPF default).
	// Full implementation maps ingress/egress selectors to identity ranges.
	return nil // TODO: full selector → identity mapping
}
