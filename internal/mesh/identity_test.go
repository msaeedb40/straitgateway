package mesh

import (
	"net"
	"testing"
)

func TestIdentityEngine(t *testing.T) {
	ie := NewIdentityEngine()

	labels := map[string]string{"app": "frontend", "env": "prod"}

	// 1. GetOrCreateIdentity
	ident1 := ie.GetOrCreateIdentity("prod", "frontend-sa", labels)
	if ident1 == nil || ident1.ID < IdentityMinUser {
		t.Fatalf("expected valid identity with ID >= %d, got %v", IdentityMinUser, ident1)
	}

	// 2. Determinism: same labels must return same ID
	ident2 := ie.GetOrCreateIdentity("prod", "frontend-sa", labels)
	if ident1.ID != ident2.ID {
		t.Errorf("expected deterministic ID %d, got %d", ident1.ID, ident2.ID)
	}

	// 3. Different labels/namespace must return different ID
	otherLabels := map[string]string{"app": "backend", "env": "prod"}
	ident3 := ie.GetOrCreateIdentity("prod", "backend-sa", otherLabels)
	if ident1.ID == ident3.ID {
		t.Errorf("expected different IDs for frontend and backend, got %d", ident1.ID)
	}

	// 4. IP binding
	podIP := net.ParseIP("10.244.1.42")
	ie.BindIP(podIP, ident1.ID)

	resolved := ie.ResolveIP(podIP)
	if resolved != ident1.ID {
		t.Errorf("expected resolved ID %d, got %d", ident1.ID, resolved)
	}

	// Unknown IP defaults to IdentityWorld
	unknownIP := net.ParseIP("8.8.8.8")
	if ie.ResolveIP(unknownIP) != IdentityWorld {
		t.Errorf("expected IdentityWorld for unknown IP, got %d", ie.ResolveIP(unknownIP))
	}

	// Unbind
	ie.UnbindIP(podIP)
	if ie.ResolveIP(podIP) != IdentityWorld {
		t.Errorf("expected IdentityWorld after unbinding, got %d", ie.ResolveIP(podIP))
	}
}
