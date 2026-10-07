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
