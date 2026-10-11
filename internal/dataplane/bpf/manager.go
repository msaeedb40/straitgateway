package bpf

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/straitgateway/straitgateway/internal/datapath/loader"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
)

// DataplaneConfig configures the eBPF Dataplane Manager
type DataplaneConfig struct {
	PinPath      string
	BPFObjectDir string
	EnableTCX    bool
	EnableXDP    bool
	EnableCgroup bool
	EnableNetkit bool
	CgroupPath   string
	MapConfig    maps.Config
}

// DefaultDataplaneConfig returns sensible defaults for Linux 6.8+ and 6.12+ LTS
func DefaultDataplaneConfig() DataplaneConfig {
	return DataplaneConfig{
		PinPath:      maps.DefaultBPFFSPath,
		BPFObjectDir: "bpf",
		EnableTCX:    true,
		EnableXDP:    true,
		EnableCgroup: true,
		EnableNetkit: true,
		CgroupPath:   "/sys/fs/cgroup",
		MapConfig:    maps.DefaultConfig(),
	}
}

// DataplaneManager is the unified Go-side controller for BPF map lifecycle,
// program loading, and kernel attachment points across all networking subsystems.
type DataplaneManager struct {
	mu         sync.RWMutex
	cfg        DataplaneConfig
	mapMgr     *maps.MapManager
	loader     *loader.Loader
	offloadMgr *loader.OffloadManager

	// Active program handles
	programs map[string]*ebpf.Program
	// Active link handles
	links []link.Link
}

// NewDataplaneManager initializes the dataplane manager and its BPF maps
func NewDataplaneManager(cfg DataplaneConfig) (*DataplaneManager, error) {
	if cfg.PinPath == "" {
		cfg.PinPath = maps.DefaultBPFFSPath
	}
	if cfg.BPFObjectDir == "" {
		cfg.BPFObjectDir = "bpf"
	}
	cfg.MapConfig.PinPath = cfg.PinPath

	mapMgr, err := maps.NewMapManager(cfg.MapConfig)
	if err != nil {
		return nil, fmt.Errorf("init map manager: %w", err)
	}

	bpfLoader := loader.NewLoader(mapMgr)
	offloadMgr := loader.NewOffloadManager()

	return &DataplaneManager{
		cfg:        cfg,
		mapMgr:     mapMgr,
		loader:     bpfLoader,
		offloadMgr: offloadMgr,
		programs:   make(map[string]*ebpf.Program),
		links:      make([]link.Link, 0),
	}, nil
}

// MapManager returns the underlying BPF MapManager
func (m *DataplaneManager) MapManager() *maps.MapManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mapMgr
}

// Loader returns the BPF program loader
func (m *DataplaneManager) Loader() *loader.Loader {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loader
}

// OffloadManager returns the SmartNIC / hardware offload manager
func (m *DataplaneManager) OffloadManager() *loader.OffloadManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.offloadMgr
}

// LoadProgram loads an eBPF program ELF from disk and registers it
func (m *DataplaneManager) LoadProgram(name, relativePath, secName string) (*ebpf.Program, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullPath := filepath.Join(m.cfg.BPFObjectDir, relativePath)
	prog, err := loader.LoadFromFile(fullPath, secName)
	if err != nil {
		return nil, fmt.Errorf("load program %s from %s: %w", name, fullPath, err)
	}

	m.programs[name] = prog
	return prog, nil
}

// AttachTCXIngress attaches a TCX ingress hook to an interface
func (m *DataplaneManager) AttachTCXIngress(ifIndex int, prog *ebpf.Program) (link.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.loader.AttachTCXIngress(ifIndex, prog)
	if err != nil {
		return nil, err
	}
	m.links = append(m.links, l)
	return l, nil
}

// AttachTCXEgress attaches a TCX egress hook to an interface
func (m *DataplaneManager) AttachTCXEgress(ifIndex int, prog *ebpf.Program) (link.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.loader.AttachTCXEgress(ifIndex, prog)
	if err != nil {
		return nil, err
	}
	m.links = append(m.links, l)
	return l, nil
}

// AttachXDP attaches an XDP hook to an interface
func (m *DataplaneManager) AttachXDP(ifIndex int, prog *ebpf.Program, flags link.XDPAttachFlags) (link.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.loader.AttachXDP(ifIndex, prog, flags)
	if err != nil {
		return nil, err
	}
	m.links = append(m.links, l)
	return l, nil
}

// AttachCgroupSocketLB attaches socket-level load balancing programs to cgroup v2
func (m *DataplaneManager) AttachCgroupSocketLB(connect4, connect6, sendmsg4, sendmsg6 *ebpf.Program) ([]link.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lks, err := m.loader.AttachCgroupSocketLB(m.cfg.CgroupPath, connect4, connect6, sendmsg4, sendmsg6)
	if err != nil {
		return nil, err
	}
	m.links = append(m.links, lks...)
	return lks, nil
}

// AttachNetkit attaches TCX hooks to a netkit primary/peer pair
func (m *DataplaneManager) AttachNetkit(primaryIndex, peerIndex int, ingressProg, egressProg *ebpf.Program) ([]link.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lks, err := m.loader.AttachNetkit(primaryIndex, peerIndex, ingressProg, egressProg)
	if err != nil {
		return nil, err
	}
	m.links = append(m.links, lks...)
	return lks, nil
}

// PinMaps pins all 13 core BPF maps to the configured bpffs directory
func (m *DataplaneManager) PinMaps() error {
	return m.mapMgr.PinAll()
}

// Close gracefully unpins/closes active BPF links, programs, and maps
func (m *DataplaneManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	for _, l := range m.links {
		if err := l.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	m.links = nil

	for name, prog := range m.programs {
		if err := prog.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close program %s: %w", name, err))
		}
	}
	m.programs = make(map[string]*ebpf.Program)

	if m.mapMgr != nil {
		if err := m.mapMgr.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing dataplane manager: %v", errs)
	}
	return nil
}
