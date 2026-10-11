package netkit

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/straitgateway/straitgateway/internal/datapath/loader"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/vishvananda/netlink"
)

// Mode defines the netkit operational mode
type Mode string

const (
	// ModeL3 operates as a pure L3 device with direct IP handoff
	ModeL3 Mode = "l3"
	// ModeL2 operates with Ethernet framing
	ModeL2 Mode = "l2"
)

// Config configures the Netkit Dataplane Manager
type Config struct {
	Mode         Mode   `json:"mode"`
	DefaultMTU   int    `json:"defaultMtu"`
	BPFObjectDir string `json:"bpfObjectDir"`
}

// DefaultConfig provides sensible defaults
func DefaultConfig() Config {
	return Config{
		Mode:         ModeL3,
		DefaultMTU:   1500,
		BPFObjectDir: "bpf",
	}
}

// NetkitDevice represents an active netkit interface pair attached to eBPF
type NetkitDevice struct {
	PrimaryName string
	PeerName    string
	PrimaryIdx  int
	PeerIdx     int
	PodIP       net.IP
	IdentityID  uint32
	Links       []link.Link
}

// Manager coordinates netkit device attachment to the eBPF dataplane
type Manager struct {
	mu          sync.RWMutex
	cfg         Config
	mapMgr      *maps.MapManager
	loader      *loader.Loader
	devices     map[string]*NetkitDevice // key: PrimaryName
	ingressProg *ebpf.Program
	egressProg  *ebpf.Program
}

// NewManager creates a new Netkit Dataplane Manager
func NewManager(cfg Config, mapMgr *maps.MapManager, bpfLoader *loader.Loader) *Manager {
	if cfg.DefaultMTU == 0 {
		cfg.DefaultMTU = 1500
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeL3
	}

	return &Manager{
		cfg:     cfg,
		mapMgr:  mapMgr,
		loader:  bpfLoader,
		devices: make(map[string]*NetkitDevice),
	}
}

// LoadPrograms loads the compiled netkit eBPF programs from disk
func (m *Manager) LoadPrograms(ingressPath, egressPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inProg, err := loader.LoadFromFile(ingressPath, "netkit/peer_ingress")
	if err != nil {
		return fmt.Errorf("load netkit peer ingress: %w", err)
	}

	outProg, err := loader.LoadFromFile(egressPath, "netkit/peer_egress")
	if err != nil {
		_ = inProg.Close()
		return fmt.Errorf("load netkit peer egress: %w", err)
	}

	m.ingressProg = inProg
	m.egressProg = outProg
	return nil
}

// SetPrograms allows setting pre-loaded eBPF programs directly
func (m *Manager) SetPrograms(ingress, egress *ebpf.Program) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ingressProg = ingress
	m.egressProg = egress
}

// ipToUint32 converts IPv4 to big-endian uint32
func ipToUint32(ip net.IP) uint32 {
	v4 := ip.To4()
	if v4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(v4)
}

// AttachPodNetkit creates or binds to a netkit pair and attaches eBPF TCX hooks
func (m *Manager) AttachPodNetkit(primaryName, peerName string, podIP net.IP, identityID uint32) (*NetkitDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	primaryLink, err := netlink.LinkByName(primaryName)
	if err != nil {
		return nil, fmt.Errorf("resolve primary link %s: %w", primaryName, err)
	}
	primaryIdx := primaryLink.Attrs().Index

	var peerIdx int
	if peerLink, err := netlink.LinkByName(peerName); err == nil {
		peerIdx = peerLink.Attrs().Index
	}

	dev := &NetkitDevice{
		PrimaryName: primaryName,
		PeerName:    peerName,
		PrimaryIdx:  primaryIdx,
		PeerIdx:     peerIdx,
		PodIP:       podIP,
		IdentityID:  identityID,
		Links:       make([]link.Link, 0),
	}

	// Attach TCX hooks if programs are loaded
	if m.loader != nil && m.ingressProg != nil {
		lks, err := m.loader.AttachNetkit(primaryIdx, peerIdx, m.ingressProg, m.egressProg)
		if err != nil {
			return nil, fmt.Errorf("attach netkit eBPF: %w", err)
		}
		dev.Links = append(dev.Links, lks...)
	}

	// Register Pod Identity if map is available
	if m.mapMgr != nil && podIP != nil && identityID > 0 {
		if ipVal := ipToUint32(podIP); ipVal != 0 {
			_ = m.mapMgr.SetIdentity(ipVal, identityID)
		}
	}

	m.devices[primaryName] = dev
	return dev, nil
}

// DetachPodNetkit unbinds eBPF hooks and deregisters pod identity
func (m *Manager) DetachPodNetkit(primaryName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev, exists := m.devices[primaryName]
	if !exists {
		return nil
	}

	for _, l := range dev.Links {
		_ = l.Close()
	}

	if m.mapMgr != nil && dev.PodIP != nil {
		if ipVal := ipToUint32(dev.PodIP); ipVal != 0 {
			_ = m.mapMgr.DeleteIdentity(ipVal)
		}
	}

	delete(m.devices, primaryName)
	return nil
}

// GetDevice retrieves an attached netkit device descriptor
func (m *Manager) GetDevice(primaryName string) (*NetkitDevice, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dev, ok := m.devices[primaryName]
	return dev, ok
}

// ListDevices returns all active netkit device attachments
func (m *Manager) ListDevices() []*NetkitDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*NetkitDevice, 0, len(m.devices))
	for _, dev := range m.devices {
		res = append(res, dev)
	}
	return res
}

// Close detaches all active links and unloads eBPF programs
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, dev := range m.devices {
		for _, l := range dev.Links {
			_ = l.Close()
		}
	}
	m.devices = make(map[string]*NetkitDevice)

	if m.ingressProg != nil {
		_ = m.ingressProg.Close()
		m.ingressProg = nil
	}
	if m.egressProg != nil {
		_ = m.egressProg.Close()
		m.egressProg = nil
	}

	return nil
}
