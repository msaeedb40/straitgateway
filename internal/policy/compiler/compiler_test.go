package compiler_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/straitgateway/straitgateway/api/v1alpha4"
	"github.com/straitgateway/straitgateway/internal/policy/compiler"
	"github.com/straitgateway/straitgateway/internal/policy/ir"
)

func TestCompiler(t *testing.T) {
	comp := compiler.NewCompiler()

	policy := &v1alpha4.StraitGatewayPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "default",
		},
		Spec: v1alpha4.StraitGatewayPolicySpec{
			Ingress: []v1alpha4.PolicyRule{
				{
					Action: "ALLOW",
					From: []v1alpha4.PolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "client"},
							},
							CIDRs: []string{"10.244.0.0/16"},
						},
					},
					Ports: []v1alpha4.PolicyPort{
						{
							Protocol: "TCP",
							Port:     8080,
						},
					},
				},
			},
		},
	}

	policySet, err := comp.Compile(policy)
	if err != nil {
		t.Fatalf("Compile() failed: %v", err)
	}

	if policySet.WorkloadID != "test-policy" {
		t.Errorf("got WorkloadID %q, want 'test-policy'", policySet.WorkloadID)
	}

	if len(policySet.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(policySet.Rules))
	}

	rule := policySet.Rules[0]
	if rule.Direction != ir.DirectionIngress {
		t.Errorf("got direction %v, want INGRESS", rule.Direction)
	}

	if rule.Action != ir.ActionAllow {
		t.Errorf("got action %v, want ALLOW", rule.Action)
	}

	if rule.Protocol != ir.ProtocolTCP {
		t.Errorf("got protocol %v, want TCP", rule.Protocol)
	}

	if len(rule.Ports) == 0 || rule.Ports[0].Start != 8080 {
		t.Errorf("got ports %v, want port 8080", rule.Ports)
	}
}
