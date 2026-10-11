package mesh

import (
	"sync"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// CompiledRule represents a single compiled eBPF policy entry
type CompiledRule struct {
	Key    types.PolicyKey
	Action types.PolicyAction
}

// PolicyCompiler translates SecurityPolicy CRDs into kernel BPF policy maps
type PolicyCompiler struct {
	mu             sync.RWMutex
	identityEngine *IdentityEngine
	mapMgr         *maps.MapManager
	activePolicies map[string][]CompiledRule // key: namespace/name
}

// NewPolicyCompiler creates a PolicyCompiler
func NewPolicyCompiler(ie *IdentityEngine, mapMgr *maps.MapManager) *PolicyCompiler {
	return &PolicyCompiler{
		identityEngine: ie,
		mapMgr:         mapMgr,
		activePolicies: make(map[string][]CompiledRule),
	}
}

// CompileAndInstall translates a SecurityPolicy into BPF rules and programs the kernel
func (pc *PolicyCompiler) CompileAndInstall(policy *v1alpha1.SecurityPolicy, targetIdentityID uint32) ([]CompiledRule, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	var compiled []CompiledRule

	// 1. Translate Ingress rules (traffic to targetIdentityID)
	for _, ingress := range policy.Spec.Ingress {
		// If ports empty, wildcard port (port 0 = all ports)
		portRules := ingress.Ports
		if len(portRules) == 0 {
			portRules = []v1alpha1.SecurityPort{{Protocol: "TCP", Port: 0}}
		}

		for _, portRule := range portRules {
			protoNum := uint8(6) // TCP
			switch portRule.Protocol {
			case "UDP":
				protoNum = 17
			case "":
				protoNum = 0
			}

			// Resolve sources
			sourceIDs := []uint32{IdentityCluster}
			if len(ingress.From) > 0 {
				sourceIDs = nil
				for _, from := range ingress.From {
					if from.ServiceAccount != "" {
						ident := pc.identityEngine.GetOrCreateIdentity(policy.Namespace, from.ServiceAccount, nil)
						sourceIDs = append(sourceIDs, ident.ID)
					}
					if from.PodSelector != nil && len(from.PodSelector.MatchLabels) > 0 {
						ident := pc.identityEngine.GetOrCreateIdentity(policy.Namespace, "default", from.PodSelector.MatchLabels)
						sourceIDs = append(sourceIDs, ident.ID)
					}
					if from.CIDR != "" {
						sourceIDs = append(sourceIDs, IdentityWorld)
					}
				}
			}

			for _, srcID := range sourceIDs {
				ruleKey := types.PolicyKey{
					SrcID:   srcID,
					DstID:   targetIdentityID,
					DstPort: uint16(portRule.Port),
					Proto:   protoNum,
				}
				action := types.PolicyAllow
				if policy.Spec.DefaultAction == v1alpha1.ActionAudit {
					action = types.PolicyAudit
				}

				compiled = append(compiled, CompiledRule{
					Key:    ruleKey,
					Action: action,
				})

				// Install into kernel BPF map if mapMgr is connected
				if pc.mapMgr != nil {
					_ = pc.mapMgr.SetPolicy(ruleKey, action)
				}
			}
		}
	}

	// 2. Translate Egress rules (traffic from targetIdentityID)
	for _, egress := range policy.Spec.Egress {
		portRules := egress.Ports
		if len(portRules) == 0 {
			portRules = []v1alpha1.SecurityPort{{Protocol: "TCP", Port: 0}}
		}

		for _, portRule := range portRules {
			protoNum := uint8(6)
			switch portRule.Protocol {
			case "UDP":
				protoNum = 17
			case "":
				protoNum = 0
			}

			destIDs := []uint32{IdentityCluster}
			if len(egress.To) > 0 {
				destIDs = nil
				for _, to := range egress.To {
					if to.ServiceAccount != "" {
						ident := pc.identityEngine.GetOrCreateIdentity(policy.Namespace, to.ServiceAccount, nil)
						destIDs = append(destIDs, ident.ID)
					}
					if to.PodSelector != nil && len(to.PodSelector.MatchLabels) > 0 {
						ident := pc.identityEngine.GetOrCreateIdentity(policy.Namespace, "default", to.PodSelector.MatchLabels)
						destIDs = append(destIDs, ident.ID)
					}
					if to.CIDR != "" {
						destIDs = append(destIDs, IdentityWorld)
					}
				}
			}

			for _, dstID := range destIDs {
				ruleKey := types.PolicyKey{
					SrcID:   targetIdentityID,
					DstID:   dstID,
					DstPort: uint16(portRule.Port),
					Proto:   protoNum,
				}
				action := types.PolicyAllow
				if policy.Spec.DefaultAction == v1alpha1.ActionAudit {
					action = types.PolicyAudit
				}

				compiled = append(compiled, CompiledRule{
					Key:    ruleKey,
					Action: action,
				})

				if pc.mapMgr != nil {
					_ = pc.mapMgr.SetPolicy(ruleKey, action)
				}
			}
		}
	}

	policyKey := policy.Namespace + "/" + policy.Name
	pc.activePolicies[policyKey] = compiled
	return compiled, nil
}

// RemovePolicy cleans up rules from kernel maps when a SecurityPolicy is deleted
func (pc *PolicyCompiler) RemovePolicy(namespace, name string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	policyKey := namespace + "/" + name
	rules, exists := pc.activePolicies[policyKey]
	if !exists {
		return nil
	}

	if pc.mapMgr != nil {
		for _, r := range rules {
			_ = pc.mapMgr.DeletePolicy(r.Key)
		}
	}

	delete(pc.activePolicies, policyKey)
	return nil
}

// GetRules returns compiled rules for a policy
func (pc *PolicyCompiler) GetRules(namespace, name string) ([]CompiledRule, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	policyKey := namespace + "/" + name
	rules, exists := pc.activePolicies[policyKey]
	return rules, exists
}

// ListPolicies returns all active compiled policies
func (pc *PolicyCompiler) ListPolicies() map[string][]CompiledRule {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make(map[string][]CompiledRule, len(pc.activePolicies))
	for k, v := range pc.activePolicies {
		result[k] = v
	}
	return result
}

// GetActiveRuleCount returns the total number of currently active rules across all policies
func (pc *PolicyCompiler) GetActiveRuleCount() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	total := 0
	for _, rules := range pc.activePolicies {
		total += len(rules)
	}
	return total
}
