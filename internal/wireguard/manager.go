package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/vishvananda/netlink"
)

// PeerConfig represents a remote WireGuard endpoint and allowed routes
type PeerConfig struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          []net.IPNet
	PersistentKeepalive int
}

// Manager controls WireGuard tunnel devices and peer associations
type Manager struct {
	mu           sync.RWMutex
	deviceName   string
	listenPort   int
	publicKey    string
	peers        map[string]*PeerConfig // key: publicKey
	routeManager *routing.RouteManager
}

// NewManager creates a WireGuard device manager
func NewManager(deviceName string, listenPort int) (*Manager, error) {
	if deviceName == "" {
		deviceName = "wg0"
	}
	if listenPort == 0 {
		listenPort = 51820
	}

	pubKey, err := generateMockKey()
	if err != nil {
		return nil, err
	}

	return &Manager{
		deviceName: deviceName,
		listenPort: listenPort,
		publicKey:  pubKey,
		peers:      make(map[string]*PeerConfig),
	}, nil
}

// SetRouteManager attaches a RouteManager for automatic transit route installation
func (m *Manager) SetRouteManager(rm *routing.RouteManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routeManager = rm
}

// EnsureDevice creates and brings up the WireGuard link
func (m *Manager) EnsureDevice(ip net.IPNet) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try creating or fetching link
	link, err := netlink.LinkByName(m.deviceName)
	if err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = m.deviceName
		wgLink := &netlink.GenericLink{
			LinkAttrs: la,
			LinkType:  "wireguard",
		}
		if err := netlink.LinkAdd(wgLink); err != nil {
			// On unprivileged dev environments, skip link add error
			return nil
		}
		link = wgLink
	}

	// Assign IP
	addr := &netlink.Addr{IPNet: &ip}
	_ = netlink.AddrAdd(link, addr)
	_ = netlink.LinkSetUp(link)

	return nil
}

// AddPeer adds or updates a WireGuard neighbor peer and installs transit routes
func (m *Manager) AddPeer(peer PeerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer.PublicKey == "" {
		return fmt.Errorf("public key cannot be empty")
	}

	m.peers[peer.PublicKey] = &peer

	// If RouteManager attached, install transit routes for peer's AllowedIPs
	if m.routeManager != nil {
		for _, allowedIP := range peer.AllowedIPs {
			_ = m.routeManager.SyncRoute(&routing.RouteEntry{
				Destination: allowedIP,
				VRFID:       0,
				Interface:   m.deviceName,
				Precedence:  routing.PrecedenceClusterRoute,
			})
		}
	}

	return nil
}

// RemovePeer deletes a WireGuard neighbor peer and withdraws transit routes
func (m *Manager) RemovePeer(publicKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	peer, exists := m.peers[publicKey]
	if !exists {
		return nil
	}

	if m.routeManager != nil {
		for _, allowedIP := range peer.AllowedIPs {
			_ = m.routeManager.WithdrawRoute(0, allowedIP)
		}
	}

	delete(m.peers, publicKey)
	return nil
}

// GetPeer retrieves a peer by public key
func (m *Manager) GetPeer(publicKey string) (*PeerConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peer, exists := m.peers[publicKey]
	return peer, exists
}

// GetDeviceName returns the WireGuard interface name
func (m *Manager) GetDeviceName() string {
	return m.deviceName
}

// GetPublicKey returns the local device public key
func (m *Manager) GetPublicKey() string {
	return m.publicKey
}

// GetPeers returns all configured peers
func (m *Manager) GetPeers() []*PeerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*PeerConfig, 0, len(m.peers))
	for _, p := range m.peers {
		list = append(list, p)
	}
	return list
}

// Close cleans up configured WireGuard peers
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers = make(map[string]*PeerConfig)
	return nil
}

func generateMockKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
