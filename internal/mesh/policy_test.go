package mesh

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/pkg/types"
)

func TestPolicyCompiler(t *testing.T) {
	ie := NewIdentityEngine()
	compiler := NewPolicyCompiler(ie, nil)

	policy := &v1alpha1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-allow",
			Namespace: "production",
		},
		Spec: v1alpha1.SecurityPolicySpec{
			DefaultAction: v1alpha1.ActionDeny,
			Ingress: []v1alpha1.RuleIngress{
				{
					From: []v1alpha1.RulePeer{
						{ServiceAccount: "frontend"},
						{CIDR: "10.0.0.0/8"},
					},
					Ports: []v1alpha1.SecurityPort{
						{Protocol: "TCP", Port: 8080},
					},
				},
			},
		},
	}

	targetID := uint32(5000)
	rules, err := compiler.CompileAndInstall(policy, targetID)
	if err != nil {
		t.Fatalf("failed to compile policy: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("expected 2 compiled rules, got %d", len(rules))
	}

	// Verify rule destination and action
	for _, r := range rules {
		if r.Key.DstID != targetID {
			t.Errorf("expected DstID %d, got %d", targetID, r.Key.DstID)
		}
		if r.Key.DstPort != 8080 {
			t.Errorf("expected DstPort 8080, got %d", r.Key.DstPort)
		}
		if r.Action != types.PolicyAllow {
			t.Errorf("expected Action PolicyAllow, got %v", r.Action)
		}
	}

	// Verify removal
	if err := compiler.RemovePolicy("production", "api-allow"); err != nil {
		t.Fatalf("remove policy failed: %v", err)
	}

	_, exists := compiler.GetRules("production", "api-allow")
	if exists {
		t.Error("expected rules to be deleted")
	}
}

func TestPolicyCompilerEgressAndAudit(t *testing.T) {
	ie := NewIdentityEngine()
	compiler := NewPolicyCompiler(ie, nil)

	policy := &v1alpha1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-egress",
			Namespace: "production",
		},
		Spec: v1alpha1.SecurityPolicySpec{
			DefaultAction: v1alpha1.ActionAudit,
			Egress: []v1alpha1.RuleEgress{
				{
					To: []v1alpha1.RulePeer{
						{ServiceAccount: "db-sa"},
					},
					Ports: []v1alpha1.SecurityPort{
						{Protocol: "TCP", Port: 5432},
					},
				},
			},
		},
	}

	targetID := uint32(6000)
	rules, err := compiler.CompileAndInstall(policy, targetID)
	if err != nil {
		t.Fatalf("failed to compile egress policy: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Key.SrcID != targetID {
		t.Errorf("expected SrcID %d, got %d", targetID, rules[0].Key.SrcID)
	}
	if rules[0].Key.DstPort != 5432 {
		t.Errorf("expected DstPort 5432, got %d", rules[0].Key.DstPort)
	}
	if rules[0].Action != types.PolicyAudit {
		t.Errorf("expected Action PolicyAudit, got %v", rules[0].Action)
	}

	if compiler.GetActiveRuleCount() != 1 {
		t.Errorf("expected 1 active rule count, got %d", compiler.GetActiveRuleCount())
	}
	if len(compiler.ListPolicies()) != 1 {
		t.Errorf("expected 1 active policy, got %d", len(compiler.ListPolicies()))
	}
}
