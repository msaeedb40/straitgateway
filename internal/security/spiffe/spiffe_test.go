package spiffe

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"testing"
	"time"
)

func TestParseSPIFFEID(t *testing.T) {
	// Valid ID
	raw := "spiffe://cluster.local/ns/production/sa/frontend-sa"
	id, err := ParseID(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing valid SPIFFE ID: %v", err)
	}
	if id.TrustDomain != "cluster.local" {
		t.Errorf("expected trust domain cluster.local, got %s", id.TrustDomain)
	}
	if id.Namespace != "production" {
		t.Errorf("expected namespace production, got %s", id.Namespace)
	}
	if id.ServiceAccount != "frontend-sa" {
		t.Errorf("expected service account frontend-sa, got %s", id.ServiceAccount)
	}

	// Invalid scheme
	_, err = ParseID("https://cluster.local/ns/default")
	if err == nil {
		t.Errorf("expected error for non-spiffe scheme")
	}

	// Missing trust domain
	_, err = ParseID("spiffe:///ns/default")
	if err == nil {
		t.Errorf("expected error for missing trust domain")
	}
}

func TestAttestorWorkloadRegistration(t *testing.T) {
	attestor := NewAttestor("cluster.local")

	// 1. Register identity
	id, err := attestor.RegisterWorkloadIdentity("spiffe://cluster.local/ns/prod/sa/payment-sa", 1005)
	if err != nil {
		t.Fatalf("failed to register workload identity: %v", err)
	}
	if id.ServiceAccount != "payment-sa" {
		t.Errorf("expected payment-sa, got %s", id.ServiceAccount)
	}

	// 2. Query token
	token, exists := attestor.GetSecurityToken("spiffe://cluster.local/ns/prod/sa/payment-sa")
	if !exists || token != 1005 {
		t.Errorf("expected security token 1005, got %d", token)
	}

	// 3. Mismatched trust domain
	_, err = attestor.RegisterWorkloadIdentity("spiffe://remote.domain/ns/prod/sa/sa", 1006)
	if err == nil {
		t.Errorf("expected error registering mismatched trust domain")
	}
}

func TestAttestorValidateSVID(t *testing.T) {
	attestor := NewAttestor("cluster.local")

	// Generate self-signed certificate with SPIFFE SAN URI
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	spiffeURI, _ := url.Parse("spiffe://cluster.local/ns/default/sa/backend-sa")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "backend-sa",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(1 * time.Hour),
		URIs:      []*url.URL{spiffeURI},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	svid, err := attestor.ValidateSVID(cert)
	if err != nil {
		t.Fatalf("failed to validate SVID: %v", err)
	}

	if svid.ID.ServiceAccount != "backend-sa" {
		t.Errorf("expected backend-sa SVID, got %s", svid.ID.ServiceAccount)
	}
	if attestor.ActiveSVIDsCount() != 1 {
		t.Errorf("expected 1 cached active SVID, got %d", attestor.ActiveSVIDsCount())
	}
}
