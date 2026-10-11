package policy

import (
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/internal/security/fqdn"
	"github.com/straitgateway/straitgateway/internal/security/ktls"
	"github.com/straitgateway/straitgateway/internal/security/spiffe"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// PolicyState tracks compiled rules and identity bindings
type PolicyState struct {
	Namespace string
	Name      string
	Rules     []mesh.CompiledRule
}

// Manager coordinates security policy compilation and kernel enforcement
type Manager struct {
	mu             sync.RWMutex
	mapMgr         *maps.MapManager
	identityEngine *mesh.IdentityEngine
	compiler       *mesh.PolicyCompiler
	policies       map[string]*PolicyState // key: namespace/name
	spiffeAttestor *spiffe.Attestor
	ktlsMgr        *ktls.Manager
	fqdnEngine     *fqdn.Engine
}

// NewManager creates a new Policy Dataplane Manager
func NewManager(mapMgr *maps.MapManager) *Manager {
	ie := mesh.NewIdentityEngine()
	if mapMgr != nil {
		ie.SetMapManager(mapMgr)
	}
	compiler := mesh.NewPolicyCompiler(ie, mapMgr)

	attestor := spiffe.NewAttestor("cluster.local")
	ktlsManager := ktls.NewManager()
	fqdnEngine := fqdn.NewEngine()

	return &Manager{
		mapMgr:         mapMgr,
		identityEngine: ie,
		compiler:       compiler,
		policies:       make(map[string]*PolicyState),
		spiffeAttestor: attestor,
		ktlsMgr:        ktlsManager,
		fqdnEngine:     fqdnEngine,
	}
}

// IdentityEngine returns the underlying workload identity engine
func (m *Manager) IdentityEngine() *mesh.IdentityEngine {
	return m.identityEngine
}

// Compiler returns the underlying policy compiler
func (m *Manager) Compiler() *mesh.PolicyCompiler {
	return m.compiler
}

// RegisterWorkloadIdentity derives a security token and binds an IP address
func (m *Manager) RegisterWorkloadIdentity(namespace, serviceAccount string, labels map[string]string, ip net.IP) uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	ident := m.identityEngine.GetOrCreateIdentity(namespace, serviceAccount, labels)
	if ip != nil {
		m.identityEngine.BindIP(ip, ident.ID)
	}
	return ident.ID
}

// DeregisterWorkloadIP unbinds an IP address from the security identity
func (m *Manager) DeregisterWorkloadIP(ip net.IP) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ip != nil {
		m.identityEngine.UnbindIP(ip)
	}
}

// InstallSecurityPolicy compiles a SecurityPolicy CRD and installs rules into the dataplane
func (m *Manager) InstallSecurityPolicy(policy *v1alpha1.SecurityPolicy, targetIdentityID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy == nil {
		return fmt.Errorf("security policy cannot be nil")
	}

	rules, err := m.compiler.CompileAndInstall(policy, targetIdentityID)
	if err != nil {
		return fmt.Errorf("compile security policy: %w", err)
	}

	key := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
	m.policies[key] = &PolicyState{
		Namespace: policy.Namespace,
		Name:      policy.Name,
		Rules:     rules,
	}

	return nil
}

// DeleteSecurityPolicy purges a SecurityPolicy from the eBPF dataplane
func (m *Manager) DeleteSecurityPolicy(namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	if _, exists := m.policies[key]; !exists {
		return nil
	}

	err := m.compiler.RemovePolicy(namespace, name)
	delete(m.policies, key)
	return err
}

// LookupVerdict evaluates a packet key against active kernel policy rules
func (m *Manager) LookupVerdict(srcID, dstID uint32, dstPort uint16, proto uint8) (types.PolicyAction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mapMgr == nil {
		return types.PolicyAllow, nil
	}

	key := types.PolicyKey{
		SrcID:   srcID,
		DstID:   dstID,
		DstPort: dstPort,
		Proto:   proto,
	}

	// 1. Exact match
	action, err := m.mapMgr.LookupPolicy(key)
	if err == nil {
		return action, nil
	}

	// 2. Wildcard port match (port 0)
	key.DstPort = 0
	action, err = m.mapMgr.LookupPolicy(key)
	if err == nil {
		return action, nil
	}

	// 3. Wildcard proto match (proto 0)
	key.Proto = 0
	action, err = m.mapMgr.LookupPolicy(key)
	if err == nil {
		return action, nil
	}

	// Default-deny for protected destination identities
	return types.PolicyDrop, nil
}

// SPIFFE returns the workload attestor
func (m *Manager) SPIFFE() *spiffe.Attestor {
	return m.spiffeAttestor
}

// KTLS returns the in-kernel TLS manager
func (m *Manager) KTLS() *ktls.Manager {
	return m.ktlsMgr
}

// FQDN returns the egress domain firewall engine
func (m *Manager) FQDN() *fqdn.Engine {
	return m.fqdnEngine
}

// ConfigureSocketEncryption sets up kTLS on a TCP socket FD
func (m *Manager) ConfigureSocketEncryption(sessionID string, socketFD int, tx, rx *ktls.CryptoInfo, hwOffload bool) (*ktls.SessionState, error) {
	return m.ktlsMgr.ConfigureSocketKTLS(sessionID, socketFD, tx, rx, hwOffload)
}

// AuthorizeEgressIP verifies whether an outbound IP connection is permitted by FQDN policies
func (m *Manager) AuthorizeEgressIP(targetIP net.IP) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fqdnEngine.IsIPAllowed(targetIP)
}
