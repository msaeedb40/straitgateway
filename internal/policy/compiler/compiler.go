// Package compiler compiles Kubernetes NetworkPolicy and StraitGatewayPolicy CRDs
// into canonical Policy Intermediate Representation (IR).
package compiler

import (
	"fmt"

	"github.com/straitgateway/straitgateway/api/v1alpha4"
	"github.com/straitgateway/straitgateway/internal/policy/cel"
	"github.com/straitgateway/straitgateway/internal/policy/ir"
)

// Compiler compiles high-level policies into canonical IR rules.
type Compiler struct {
	celEvaluator *cel.Evaluator
}

// NewCompiler creates a new policy compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		celEvaluator: cel.NewEvaluator(),
	}
}

// Compile compiles a StraitGatewayPolicy CRD into a set of IR rules.
func (c *Compiler) Compile(policy *v1alpha4.StraitGatewayPolicy) (*ir.PolicySet, error) {
	if policy == nil {
		return nil, fmt.Errorf("policy cannot be nil")
	}

	policySet := &ir.PolicySet{
		WorkloadID: policy.Name,
		Rules:      make([]ir.Rule, 0),
	}

	// Compile Ingress rules
	for i, ing := range policy.Spec.Ingress {
		rule := ir.Rule{
			ID:        fmt.Sprintf("%s-ingress-%d", policy.Name, i),
			Name:      fmt.Sprintf("%s-in-%d", policy.Name, i),
			Direction: ir.DirectionIngress,
			Action:    ir.Action(ing.Action),
			Priority:  100,
		}

		if rule.Action == "" {
			rule.Action = ir.ActionAllow
		}

		for _, from := range ing.From {
			if from.PodSelector != nil && len(from.PodSelector.MatchLabels) > 0 {
				rule.Source.Labels = from.PodSelector.MatchLabels
			}
			if from.NamespaceSelector != nil && len(from.NamespaceSelector.MatchLabels) > 0 {
				rule.Source.Namespaces = []string{from.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]}
			}
			if len(from.CIDRs) > 0 {
				rule.Source.CIDRs = append(rule.Source.CIDRs, from.CIDRs...)
			}
		}

		for _, port := range ing.Ports {
			rule.Protocol = ir.Protocol(port.Protocol)
			if port.Port > 0 {
				rule.Ports = append(rule.Ports, ir.PortRange{
					Start: uint16(port.Port),
					End:   uint16(port.Port),
				})
			}
		}

		policySet.Rules = append(policySet.Rules, rule)
	}

	// Compile Egress rules
	for i, eg := range policy.Spec.Egress {
		rule := ir.Rule{
			ID:        fmt.Sprintf("%s-egress-%d", policy.Name, i),
			Name:      fmt.Sprintf("%s-eg-%d", policy.Name, i),
			Direction: ir.DirectionEgress,
			Action:    ir.Action(eg.Action),
			Priority:  100,
		}

		if rule.Action == "" {
			rule.Action = ir.ActionAllow
		}

		for _, to := range eg.To {
			if to.PodSelector != nil && len(to.PodSelector.MatchLabels) > 0 {
				rule.Destination.Labels = to.PodSelector.MatchLabels
			}
			if to.NamespaceSelector != nil && len(to.NamespaceSelector.MatchLabels) > 0 {
				rule.Destination.Namespaces = []string{to.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]}
			}
			if len(to.CIDRs) > 0 {
				rule.Destination.CIDRs = append(rule.Destination.CIDRs, to.CIDRs...)
			}
		}

		for _, port := range eg.Ports {
			rule.Protocol = ir.Protocol(port.Protocol)
			if port.Port > 0 {
				rule.Ports = append(rule.Ports, ir.PortRange{
					Start: uint16(port.Port),
					End:   uint16(port.Port),
				})
			}
		}

		policySet.Rules = append(policySet.Rules, rule)
	}

	return policySet, nil
}
