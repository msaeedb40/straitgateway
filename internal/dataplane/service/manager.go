package service

import (
	"fmt"
	"net"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/straitgateway/straitgateway/internal/datapath/loader"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/loadbalancer"
)

// Config holds configuration for the Service Dataplane Manager
type Config struct {
	CgroupPath   string `json:"cgroupPath"`
	BPFObjectDir string `json:"bpfObjectDir"`
	EnableIPv6   bool   `json:"enableIPv6"`
}

// DefaultConfig provides standard service dataplane configuration
func DefaultConfig() Config {
	return Config{
		CgroupPath:   "/sys/fs/cgroup",
		BPFObjectDir: "bpf",
		EnableIPv6:   false,
	}
}

// Manager coordinates socket-level and TC-level service load balancing
type Manager struct {
	mu        sync.RWMutex
	cfg       Config
	mapMgr    *maps.MapManager
	bpfLoader *loader.Loader
	lbMgr     *loadbalancer.ServiceManager
	con4Prog  *ebpf.Program
	con6Prog  *ebpf.Program
	snd4Prog  *ebpf.Program
	snd6Prog  *ebpf.Program
	tcLBProg  *ebpf.Program
	links     []link.Link
}

// NewManager creates a new Service Dataplane Manager
func NewManager(cfg Config, mapMgr *maps.MapManager, bpfLoader *loader.Loader) *Manager {
	if cfg.CgroupPath == "" {
		cfg.CgroupPath = "/sys/fs/cgroup"
	}
	if cfg.BPFObjectDir == "" {
		cfg.BPFObjectDir = "bpf"
	}

	lbMgr := loadbalancer.NewServiceManager(mapMgr)

	return &Manager{
		cfg:       cfg,
		mapMgr:    mapMgr,
		bpfLoader: bpfLoader,
		lbMgr:     lbMgr,
		links:     make([]link.Link, 0),
	}
}

// LoadPrograms loads compiled cgroup socket LB and TC service LB BPF programs
func (m *Manager) LoadPrograms(con4Path, con6Path, snd4Path, snd6Path, tcLBPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	if con4Path != "" {
		m.con4Prog, err = loader.LoadFromFile(con4Path, "cgroup/connect4")
		if err != nil {
			return fmt.Errorf("load cgroup connect4: %w", err)
		}
	}

	if con6Path != "" && m.cfg.EnableIPv6 {
		m.con6Prog, err = loader.LoadFromFile(con6Path, "cgroup/connect6")
		if err != nil {
			return fmt.Errorf("load cgroup connect6: %w", err)
		}
	}

	if snd4Path != "" {
		m.snd4Prog, err = loader.LoadFromFile(snd4Path, "cgroup/sendmsg4")
		if err != nil {
			return fmt.Errorf("load cgroup sendmsg4: %w", err)
		}
	}

	if snd6Path != "" && m.cfg.EnableIPv6 {
		m.snd6Prog, err = loader.LoadFromFile(snd6Path, "cgroup/sendmsg6")
		if err != nil {
			return fmt.Errorf("load cgroup sendmsg6: %w", err)
		}
	}

	if tcLBPath != "" {
		m.tcLBProg, err = loader.LoadFromFile(tcLBPath, "tc/service_lb")
		if err != nil {
			return fmt.Errorf("load tc service_lb: %w", err)
		}
	}

	return nil
}

// AttachSocketLB attaches loaded socket translation programs to cgroup v2
func (m *Manager) AttachSocketLB() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.bpfLoader == nil {
		return fmt.Errorf("bpf loader not configured")
	}

	lks, err := m.bpfLoader.AttachCgroupSocketLB(
		m.cfg.CgroupPath,
		m.con4Prog,
		m.con6Prog,
		m.snd4Prog,
		m.snd6Prog,
	)
	if err != nil {
		return fmt.Errorf("attach socket LB to cgroup %s: %w", m.cfg.CgroupPath, err)
	}

	m.links = append(m.links, lks...)
	return nil
}

// ServiceManager returns the underlying Go service VIP / Maglev manager
func (m *Manager) ServiceManager() *loadbalancer.ServiceManager {
	return m.lbMgr
}

// RegisterService adds a service and populates BPF maps with Maglev backend selection
func (m *Manager) RegisterService(namespace, name string, vip net.IP, port uint16, proto uint8, backends []loadbalancer.Backend) error {
	svc := &loadbalancer.ServiceEntry{
		Namespace: namespace,
		Name:      name,
		VIP:       vip,
		Port:      port,
		Protocol:  proto,
		Backends:  backends,
	}
	return m.lbMgr.UpsertService(svc)
}

// UnregisterService removes a service VIP and cleans up BPF maps
func (m *Manager) UnregisterService(namespace, name string, port uint16) error {
	return m.lbMgr.DeleteService(namespace, name, port)
}

// Close detaches links and unloads BPF programs
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, l := range m.links {
		_ = l.Close()
	}
	m.links = nil

	if m.con4Prog != nil {
		_ = m.con4Prog.Close()
	}
	if m.con6Prog != nil {
		_ = m.con6Prog.Close()
	}
	if m.snd4Prog != nil {
		_ = m.snd4Prog.Close()
	}
	if m.snd6Prog != nil {
		_ = m.snd6Prog.Close()
	}
	if m.tcLBProg != nil {
		_ = m.tcLBProg.Close()
	}

	return nil
}
