package security

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/mesh"
)

func TestSecurityPolicyReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	policy := &v1alpha1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "default",
		},
		Spec: v1alpha1.SecurityPolicySpec{
			DefaultAction: v1alpha1.ActionDeny,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "payment"},
			},
			Ingress: []v1alpha1.RuleIngress{
				{
					From: []v1alpha1.RulePeer{
						{ServiceAccount: "order-service"},
					},
					Ports: []v1alpha1.SecurityPort{
						{Protocol: "TCP", Port: 8443},
					},
				},
			},
		},
	}

	fakeClient := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(policy).Build()
	ie := mesh.NewIdentityEngine()
	compiler := mesh.NewPolicyCompiler(ie, nil)

	r := &SecurityPolicyReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Compiler: compiler,
		Identity: ie,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "test-policy",
		},
	}

	// 1. Reconcile Policy creation
	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if res.Requeue {
		t.Error("expected no requeue")
	}

	rules, found := compiler.GetRules("default", "test-policy")
	if !found || len(rules) != 1 {
		t.Fatalf("expected 1 rule compiled, got %d (found: %v)", len(rules), found)
	}
	if rules[0].Key.DstPort != 8443 {
		t.Errorf("expected DstPort 8443, got %d", rules[0].Key.DstPort)
	}

	// 2. Reconcile Policy deletion
	_ = fakeClient.Delete(context.Background(), policy)
	_, err = r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile on delete failed: %v", err)
	}

	_, found = compiler.GetRules("default", "test-policy")
	if found {
		t.Error("expected rules to be deleted after policy deletion")
	}
}
