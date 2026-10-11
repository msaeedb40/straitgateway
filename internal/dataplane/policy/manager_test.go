package policy

import (
	"net"
	"testing"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/internal/security/ktls"
	"github.com/straitgateway/straitgateway/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPolicyManagerLifecycle(t *testing.T) {
	mgr := NewManager(nil)
	if mgr == nil {
		t.Fatalf("expected non-nil Policy Manager")
	}

	// 1. Identity registration
	podIP := net.ParseIP("10.244.0.12")
	id := mgr.RegisterWorkloadIdentity("production", "frontend-sa", map[string]string{"app": "frontend"}, podIP)
	if id < 1000 {
		t.Errorf("expected user workload identity >= 1000, got %d", id)
	}

	retrievedID := mgr.IdentityEngine().ResolveIP(podIP)
	if retrievedID != id {
		t.Errorf("expected identity %d for IP %s, got %d", id, podIP, retrievedID)
	}

	// 2. Policy installation
	policy := &v1alpha1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-frontend",
			Namespace: "production",
		},
		Spec: v1alpha1.SecurityPolicySpec{
			DefaultAction: v1alpha1.ActionDeny,
			Ingress: []v1alpha1.RuleIngress{
				{
					Ports: []v1alpha1.SecurityPort{
						{
							Protocol: "TCP",
							Port:     8080,
						},
					},
					From: []v1alpha1.RulePeer{
						{
							ServiceAccount: "frontend-sa",
						},
					},
				},
			},
		},
	}

	targetID := uint32(2000)
	if err := mgr.InstallSecurityPolicy(policy, targetID); err != nil {
		t.Fatalf("failed to install security policy: %v", err)
	}

	// 3. Verdict verification (in memory mode without BPF map returns Allow)
	verdict, err := mgr.LookupVerdict(id, targetID, 8080, 6)
	if err != nil {
		t.Errorf("unexpected error on lookup: %v", err)
	}
	if verdict != types.PolicyAllow {
		t.Errorf("expected PolicyAllow without map manager, got %v", verdict)
	}

	// 4. Policy deletion
	if err := mgr.DeleteSecurityPolicy("production", "allow-frontend"); err != nil {
		t.Errorf("failed to delete security policy: %v", err)
	}

	// 5. IP deregistration
	mgr.DeregisterWorkloadIP(podIP)
	if unmapped := mgr.IdentityEngine().ResolveIP(podIP); unmapped != mesh.IdentityWorld {
		t.Errorf("expected IP to be resolved as IdentityWorld, got %d", unmapped)
	}
}

func TestEnterpriseZeroTrustIntegration(t *testing.T) {
	mgr := NewManager(nil)

	// 1. SPIFFE Attestation
	id, err := mgr.SPIFFE().RegisterWorkloadIdentity("spiffe://cluster.local/ns/finance/sa/payment-sa", 1050)
	if err != nil {
		t.Fatalf("failed to register SPIFFE identity: %v", err)
	}
	if id.Namespace != "finance" || id.ServiceAccount != "payment-sa" {
		t.Errorf("parsed SPIFFE ID mismatch: %+v", id)
	}

	token, ok := mgr.SPIFFE().GetSecurityToken("spiffe://cluster.local/ns/finance/sa/payment-sa")
	if !ok || token != 1050 {
		t.Errorf("expected token 1050, got %d", token)
	}

	// 2. kTLS In-Kernel Session Setup
	crypto := ktls.CryptoInfo{
		Cipher:  ktls.CipherAESGCM256,
		Version: ktls.TLSVersion13,
		Key:     make([]byte, 32),
		IV:      make([]byte, 12),
	}
	session, err := mgr.ConfigureSocketEncryption("sess-1", 15, &crypto, &crypto, false)
	if err != nil {
		t.Fatalf("failed to configure kTLS: %v", err)
	}
	if session.Cipher != ktls.CipherAESGCM256 {
		t.Errorf("expected CipherAESGCM256, got %s", session.Cipher)
	}

	// 3. FQDN Dynamic Egress Firewall
	_, err = mgr.FQDN().RegisterRule("*.api.stripe.com", "ALLOW")
	if err != nil {
		t.Fatalf("failed to register FQDN rule: %v", err)
	}
	stripeIP := net.ParseIP("54.187.12.34")
	_, _ = mgr.FQDN().ProcessDNSResponse("checkout.api.stripe.com", []net.IP{stripeIP}, 60)

	if !mgr.AuthorizeEgressIP(stripeIP) {
		t.Errorf("expected stripe IP %s to be authorized by FQDN engine", stripeIP)
	}

	unauthorizedIP := net.ParseIP("198.51.100.99")
	if mgr.AuthorizeEgressIP(unauthorizedIP) {
		t.Errorf("unauthorized IP should not be permitted by FQDN engine")
	}
}
