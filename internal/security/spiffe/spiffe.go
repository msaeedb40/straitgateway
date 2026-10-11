package spiffe

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ID represents a validated SPIFFE ID (e.g., spiffe://cluster.local/ns/default/sa/frontend-sa)
type ID struct {
	TrustDomain    string `json:"trustDomain"`
	Namespace      string `json:"namespace"`
	ServiceAccount string `json:"serviceAccount"`
	Raw            string `json:"raw"`
}

// ParseID parses and validates a SPIFFE URI string
func ParseID(raw string) (*ID, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("malformed SPIFFE ID: %w", err)
	}
	if u.Scheme != "spiffe" {
		return nil, fmt.Errorf("invalid SPIFFE scheme: expected 'spiffe', got '%s'", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("missing trust domain in SPIFFE ID: %s", raw)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	ns := ""
	sa := ""

	for i := 0; i < len(parts)-1; i += 2 {
		if parts[i] == "ns" {
			ns = parts[i+1]
		} else if parts[i] == "sa" {
			sa = parts[i+1]
		}
	}

	return &ID{
		TrustDomain:    u.Host,
		Namespace:      ns,
		ServiceAccount: sa,
		Raw:            raw,
	}, nil
}

// String returns the canonical SPIFFE URI representation
func (id *ID) String() string {
	return id.Raw
}

// X509SVID represents a SPIFFE Verifiable Identity Document
type X509SVID struct {
	ID        *ID               `json:"id"`
	Cert      *x509.Certificate `json:"-"`
	ExpiresAt time.Time         `json:"expiresAt"`
	IssuedAt  time.Time         `json:"issuedAt"`
	SANs      []string          `json:"sans"`
}

// Attestor manages workload SPIFFE identity attestation and SVID verification
type Attestor struct {
	mu           sync.RWMutex
	trustDomain  string
	trustedRoots *x509.CertPool
	identities   map[string]uint32 // SPIFFE ID -> 32-bit eBPF Security Identity token
	svids        map[string]*X509SVID
}

// NewAttestor creates a SPIFFE identity attestor
func NewAttestor(trustDomain string) *Attestor {
	if trustDomain == "" {
		trustDomain = "cluster.local"
	}
	return &Attestor{
		trustDomain:  trustDomain,
		trustedRoots: x509.NewCertPool(),
		identities:   make(map[string]uint32),
		svids:        make(map[string]*X509SVID),
	}
}

// AddTrustRoot adds a trusted CA root certificate PEM
func (a *Attestor) AddTrustRoot(pemCert []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	block, _ := pem.Decode(pemCert)
	if block == nil {
		return fmt.Errorf("failed to parse PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid CA certificate: %w", err)
	}
	a.trustedRoots.AddCert(cert)
	return nil
}

// RegisterWorkloadIdentity binds a SPIFFE ID to an eBPF 32-bit security identity token
func (a *Attestor) RegisterWorkloadIdentity(spiffeID string, securityToken uint32) (*ID, error) {
	id, err := ParseID(spiffeID)
	if err != nil {
		return nil, err
	}
	if id.TrustDomain != a.trustDomain {
		return nil, fmt.Errorf("trust domain mismatch: expected %s, got %s", a.trustDomain, id.TrustDomain)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.identities[id.Raw] = securityToken
	return id, nil
}

// GetSecurityToken resolves a SPIFFE ID into its 32-bit eBPF security identity token
func (a *Attestor) GetSecurityToken(spiffeID string) (uint32, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	token, exists := a.identities[spiffeID]
	return token, exists
}

// ValidateSVID verifies certificate authenticity and validates SPIFFE SAN URI
func (a *Attestor) ValidateSVID(cert *x509.Certificate) (*X509SVID, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate cannot be nil")
	}

	now := time.Now()
	if now.Before(cert.NotBefore) {
		return nil, fmt.Errorf("certificate not yet valid")
	}
	if now.After(cert.NotAfter) {
		return nil, fmt.Errorf("certificate expired at %v", cert.NotAfter)
	}

	var spiffeURI *url.URL
	for _, u := range cert.URIs {
		if u.Scheme == "spiffe" {
			spiffeURI = u
			break
		}
	}
	if spiffeURI == nil {
		return nil, fmt.Errorf("certificate has no spiffe:// SAN URI")
	}

	id, err := ParseID(spiffeURI.String())
	if err != nil {
		return nil, err
	}

	svid := &X509SVID{
		ID:        id,
		Cert:      cert,
		IssuedAt:  cert.NotBefore,
		ExpiresAt: cert.NotAfter,
		SANs:      cert.DNSNames,
	}

	a.mu.Lock()
	a.svids[id.Raw] = svid
	a.mu.Unlock()

	return svid, nil
}

// ActiveSVIDsCount returns number of cached valid workload identities
func (a *Attestor) ActiveSVIDsCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.svids)
}
