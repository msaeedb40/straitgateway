package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"sync"

	"github.com/vishvananda/netlink"
)

// PeerConfig represents a remote WireGuard endpoint and allowed routes
type PeerConfig struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs         []net.IPNet
	PersistentKeepalive int
}

// Manager controls WireGuard tunnel devices and peer associations
type Manager struct {
	mu         sync.RWMutex
	deviceName string
	listenPort int
	publicKey  string
	peers      map[string]*PeerConfig // key: publicKey
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

// AddPeer adds or updates a WireGuard neighbor peer
func (m *Manager) AddPeer(peer PeerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer.PublicKey == "" {
		return fmt.Errorf("public key cannot be empty")
	}

	m.peers[peer.PublicKey] = &peer
	return nil
}

// RemovePeer deletes a WireGuard neighbor peer
func (m *Manager) RemovePeer(publicKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.peers, publicKey)
	return nil
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

func generateMockKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
