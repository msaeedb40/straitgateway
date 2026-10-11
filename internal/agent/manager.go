package agent

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/loadbalancer"
	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/internal/observability"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

// Config defines Agent configuration
type Config struct {
	NodeName        string
	NodeIP          net.IP
	LocalASN        uint32
	RouterID        net.IP
	WireGuardPort   int
	WireGuardDev    string
	BPFPinPath      string
	HealthInterval  time.Duration
	MetricsBindAddr string
}

// Manager coordinates all node-level networking subsystems
type Manager struct {
	mu        sync.RWMutex
	cfg       Config
	mapMgr    *maps.MapManager
	fib       *routing.FIBEngine
	router    *routing.RouteManager
	bgp       *bgp.Speaker
	bfd       *bfd.Monitor
	wg        *wireguard.Manager
	lb        *loadbalancer.ServiceManager
	ident     *mesh.IdentityEngine
	policy    *mesh.PolicyCompiler
	metrics   *observability.MetricsCollector
	flows     *observability.FlowCollector
	obsServer *observability.Server
	health    *HealthChecker
	stopCh    chan struct{}
}

// NewManager creates an Agent Manager
func NewManager(cfg Config) (*Manager, error) {
	if cfg.WireGuardDev == "" {
		cfg.WireGuardDev = "wg0"
	}
	if cfg.WireGuardPort == 0 {
		cfg.WireGuardPort = 51820
	}
	if cfg.LocalASN == 0 {
		cfg.LocalASN = 65001
	}
	if cfg.HealthInterval == 0 {
		cfg.HealthInterval = 10 * time.Second
	}

	// 1. Initialize BPF MapManager
	mapCfg := maps.DefaultConfig()
	if cfg.BPFPinPath != "" {
		mapCfg.PinPath = cfg.BPFPinPath
	}
	mapMgr, _ := maps.NewMapManager(mapCfg)

	// 2. Initialize Routing FIB
	fib := routing.NewFIBEngine()
	router := routing.NewRouteManager(fib, mapMgr)

	// 3. Initialize BGP & BFD
	rID := cfg.RouterID
	if rID == nil {
		rID = cfg.NodeIP
	}
	bgpSpeaker := bgp.NewSpeaker(cfg.LocalASN, rID, router)
	bfdMon := bfd.NewMonitor()
	bgpSpeaker.SetBFDMonitor(bfdMon)

	// 4. Initialize WireGuard
	wgMgr, _ := wireguard.NewManager(cfg.WireGuardDev, cfg.WireGuardPort)
	if wgMgr != nil {
		wgMgr.SetRouteManager(router)
	}

	// 5. Initialize Service LB and Security
	lbMgr := loadbalancer.NewServiceManager(mapMgr)
	identEngine := mesh.NewIdentityEngine()
	policyCompiler := mesh.NewPolicyCompiler(identEngine, mapMgr)

	// 6. Initialize Observability
	metricsCol := observability.NewMetricsCollector(mapMgr, bgpSpeaker, bfdMon, wgMgr, cfg.NodeName)
	flowCol := observability.NewFlowCollector(1000, identEngine, cfg.NodeName)
	var obsSrv *observability.Server
	if cfg.MetricsBindAddr != "" {
		obsSrv = observability.NewServer(cfg.MetricsBindAddr, metricsCol, flowCol, nil)
	}

	// 7. Initialize Health Checker
	healthChecker := NewHealthChecker()

	return &Manager{
		cfg:       cfg,
		mapMgr:    mapMgr,
		fib:       fib,
		router:    router,
		bgp:       bgpSpeaker,
		bfd:       bfdMon,
		wg:        wgMgr,
		lb:        lbMgr,
		ident:     identEngine,
		policy:    policyCompiler,
		metrics:   metricsCol,
		flows:     flowCol,
		obsServer: obsSrv,
		health:    healthChecker,
		stopCh:    make(chan struct{}),
	}, nil
}

// GetFIB returns the agent's forwarding information base
func (m *Manager) GetFIB() *routing.FIBEngine {
	return m.fib
}

// GetRouteManager returns the active route manager
func (m *Manager) GetRouteManager() *routing.RouteManager {
	return m.router
}

// GetBGP returns the active BGP speaker
func (m *Manager) GetBGP() *bgp.Speaker {
	return m.bgp
}

// GetBFD returns the active BFD monitor
func (m *Manager) GetBFD() *bfd.Monitor {
	return m.bfd
}

// GetWireGuard returns the active WireGuard tunnel manager
func (m *Manager) GetWireGuard() *wireguard.Manager {
	return m.wg
}

// GetMetricsCollector returns the active metrics collector
func (m *Manager) GetMetricsCollector() *observability.MetricsCollector {
	return m.metrics
}

// GetFlowCollector returns the active flow telemetry collector
func (m *Manager) GetFlowCollector() *observability.FlowCollector {
	return m.flows
}

// GetObservabilityServer returns the observability HTTP server
func (m *Manager) GetObservabilityServer() *observability.Server {
	return m.obsServer
}

// Start launches the agent daemon run loop
func (m *Manager) Start(ctx context.Context) error {
	_ = ConfigureHostNetworking()
	_ = EnsureBPFFS()

	// Pin maps if mapMgr active
	if m.mapMgr != nil {
		_ = m.mapMgr.PinAll()
	}

	// Start Observability HTTP server if configured
	if m.obsServer != nil {
		go func() {
			_ = m.obsServer.Start()
		}()
	}

	ticker := time.NewTicker(m.cfg.HealthInterval)
	defer ticker.Stop()

	fmt.Printf("StraitGateway Node Agent running on %s (IP: %v, ASN: %d)\n",
		m.cfg.NodeName, m.cfg.NodeIP, m.cfg.LocalASN)

	for {
		select {
		case <-ctx.Done():
			return m.Stop()
		case <-m.stopCh:
			return m.Stop()
		case <-ticker.C:
			status := m.health.Check()
			if !status.Healthy {
				fmt.Printf("[WARN] Agent health check failed: %s\n", status.ErrorMessage)
			}
		}
	}
}

// Stop gracefully shuts down the agent
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.stopCh:
		// already stopped
		return nil
	default:
		close(m.stopCh)
	}

	if m.obsServer != nil {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.obsServer.Stop(ctxShutdown)
	}

	if m.bfd != nil {
		m.bfd.Close()
	}
	if m.mapMgr != nil {
		_ = m.mapMgr.Close()
	}
	_ = CleanupHostNetworking()
	return nil
}

// HealthStatus returns the latest health report
func (m *Manager) HealthStatus() HealthStatus {
	return m.health.Check()
}
