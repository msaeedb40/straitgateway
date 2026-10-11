package mesh

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
)

// Well-known security identities
const (
	IdentityUnknown uint32 = 0
	IdentityHost    uint32 = 1
	IdentityWorld   uint32 = 2
	IdentityCluster uint32 = 3
	IdentityKubeDNS uint32 = 4
	IdentityMinUser uint32 = 1000
)

// Identity represents workload security metadata
type Identity struct {
	ID             uint32
	Namespace      string
	ServiceAccount string
	Labels         map[string]string
}

// IdentityEngine manages workload identity allocation and IP-to-Identity mapping
type IdentityEngine struct {
	mu           sync.RWMutex
	identities   map[uint32]*Identity
	labelIndex   map[string]uint32 // normalized descriptor -> ID
	ipToIdentity map[string]uint32 // IP string -> ID
	nextID       uint32
	mapMgr       *maps.MapManager
}

// NewIdentityEngine creates an IdentityEngine
func NewIdentityEngine() *IdentityEngine {
	ie := &IdentityEngine{
		identities:   make(map[uint32]*Identity),
		labelIndex:   make(map[string]uint32),
		ipToIdentity: make(map[string]uint32),
		nextID:       IdentityMinUser,
	}

	// Register well-known identities
	ie.identities[IdentityHost] = &Identity{ID: IdentityHost, Namespace: "kube-system", ServiceAccount: "host"}
	ie.identities[IdentityWorld] = &Identity{ID: IdentityWorld, Namespace: "", ServiceAccount: "world"}
	ie.identities[IdentityCluster] = &Identity{ID: IdentityCluster, Namespace: "kube-system", ServiceAccount: "cluster"}
	ie.identities[IdentityKubeDNS] = &Identity{ID: IdentityKubeDNS, Namespace: "kube-system", ServiceAccount: "coredns"}

	return ie
}

// GetOrCreateIdentity derives or allocates a deterministic security ID for a workload
func (ie *IdentityEngine) GetOrCreateIdentity(namespace, serviceAccount string, labels map[string]string) *Identity {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	norm := normalizeLabels(namespace, serviceAccount, labels)
	if id, exists := ie.labelIndex[norm]; exists {
		return ie.identities[id]
	}

	// Deterministic 32-bit hash above IdentityMinUser
	h := sha256.Sum256([]byte(norm))
	hashVal := binary.BigEndian.Uint32(h[:4])
	id := (hashVal % 1000000) + IdentityMinUser

	ident := &Identity{
		ID:             id,
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
		Labels:         labels,
	}

	ie.identities[id] = ident
	ie.labelIndex[norm] = id
	return ident
}

// SetMapManager attaches a BPF MapManager to synchronize kernel identity maps
func (ie *IdentityEngine) SetMapManager(mm *maps.MapManager) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.mapMgr = mm
}

// BindIP maps an IP address to a security identity and syncs kernel sg_identities
func (ie *IdentityEngine) BindIP(ip net.IP, identityID uint32) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.ipToIdentity[ip.String()] = identityID

	if ie.mapMgr != nil {
		if ip4 := ip.To4(); ip4 != nil {
			ipUint := binary.BigEndian.Uint32(ip4)
			_ = ie.mapMgr.SetIdentity(ipUint, identityID)
		}
	}
}

// UnbindIP removes the IP mapping and syncs kernel sg_identities
func (ie *IdentityEngine) UnbindIP(ip net.IP) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	delete(ie.ipToIdentity, ip.String())

	if ie.mapMgr != nil {
		if ip4 := ip.To4(); ip4 != nil {
			ipUint := binary.BigEndian.Uint32(ip4)
			_ = ie.mapMgr.DeleteIdentity(ipUint)
		}
	}
}

// ResolveIP returns the security ID associated with an IP (or IdentityWorld if unknown)
func (ie *IdentityEngine) ResolveIP(ip net.IP) uint32 {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	if id, exists := ie.ipToIdentity[ip.String()]; exists {
		return id
	}
	return IdentityWorld
}

// GetIdentityByID retrieves identity details
func (ie *IdentityEngine) GetIdentityByID(id uint32) (*Identity, bool) {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	ident, exists := ie.identities[id]
	return ident, exists
}

// ListIdentities returns all registered identities
func (ie *IdentityEngine) ListIdentities() []*Identity {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	list := make([]*Identity, 0, len(ie.identities))
	for _, ident := range ie.identities {
		list = append(list, ident)
	}
	return list
}

// GetIPBindings returns a copy of all IP-to-Identity bindings
func (ie *IdentityEngine) GetIPBindings() map[string]uint32 {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	copyMap := make(map[string]uint32, len(ie.ipToIdentity))
	for ip, id := range ie.ipToIdentity {
		copyMap[ip] = id
	}
	return copyMap
}

func normalizeLabels(ns, sa string, labels map[string]string) string {
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("ns=%s;sa=%s;", ns, sa))
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s=%s;", k, labels[k]))
	}
	return b.String()
}
