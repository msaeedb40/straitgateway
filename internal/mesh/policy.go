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

	// Translate Ingress rules
	for _, ingress := range policy.Spec.Ingress {
		for _, portRule := range ingress.Ports {
			protoNum := uint8(6) // TCP
			if portRule.Protocol == "UDP" {
				protoNum = 17
			}

			// If no specific sources specified, allow from cluster
			sourceIDs := []uint32{IdentityCluster}
			if len(ingress.From) > 0 {
				sourceIDs = nil
				for _, from := range ingress.From {
					if from.ServiceAccount != "" {
						ident := pc.identityEngine.GetOrCreateIdentity(policy.Namespace, from.ServiceAccount, nil)
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
				if policy.Spec.DefaultAction == v1alpha1.ActionDeny {
					action = types.PolicyAllow // Explicitly allowed rule
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
