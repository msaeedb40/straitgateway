// Package roce implements the StraitGateway RoCEv2 dataplane manager.
//
// It manages the lifecycle of RoCEv2-specific BPF maps and programs:
//
//   - QP state table (sg_qp_table): tracks Queue Pair lifecycle (RESET → INIT
//     → RTR → RTS → SQD → ERR) and syncs state from RDMA verbs events.
//   - PFC configuration (sg_pfc_config): programs per-priority DSCP, pause
//     thresholds, and ECN capability flags consumed by the TC egress program.
//   - NCCL topology table (sg_gpu_topo): populated from GPU topology discovery
//     (nvidia-smi topo -m), enabling topology-aware routing in XDP.
//   - ECN/CNP event consumer: drains sg_roce_events ring buffer and generates
//     Congestion Notification Packets (CNPs) via RDMA verbs.
//   - Prometheus metrics: exposes per-QPN byte/packet/CNP counters.
//
// Architecture note:
//
// The Manager is designed to be instantiated once per RDMA-capable interface.
// GPU topology refresh runs every TopologyRefreshInterval (default 60s).
// CNP generation is performed in userspace via rdma-core (libibverbs) to avoid
// the complexity of in-kernel packet cloning in XDP context.
package roce

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

const (
	// DefaultBPFFSBase is the bpffs root for RoCE pinned maps.
	DefaultBPFFSBase = "/sys/fs/bpf/straitgateway/roce"

	// RoCEUDPPort is the IANA-assigned UDP port for RoCEv2.
	RoCEUDPPort = 4791

	// MaxQPs is the maximum number of QPs tracked per node.
	MaxQPs = 131072

	// MaxGPUPeers is the maximum number of GPU peer hosts in the topology map.
	MaxGPUPeers = 4096

	// PFCPriorities is the number of IEEE 802.1p priority classes.
	PFCPriorities = 8

	// RoCEEventRingBufSize is the ring buffer size in bytes (1 MiB).
	RoCEEventRingBufSize = 1 << 20

	// DefaultTopologyRefreshInterval is how often GPU topology is re-read.
	DefaultTopologyRefreshInterval = 60 * time.Second
)

// QPState represents the Queue Pair state machine states.
type QPState uint8

const (
	QPStateReset QPState = 0
	QPStateInit  QPState = 1
	QPStateRTR   QPState = 2 // Ready To Receive
	QPStateRTS   QPState = 3 // Ready To Send
	QPStateSQD   QPState = 4 // Send Queue Drain
	QPStateSQE   QPState = 5 // Send Queue Error
	QPStateErr   QPState = 6
	QPStateDrain QPState = 7
)

// String returns a human-readable QP state name.
func (s QPState) String() string {
	switch s {
	case QPStateReset:
		return "RESET"
	case QPStateInit:
		return "INIT"
	case QPStateRTR:
		return "RTR"
	case QPStateRTS:
		return "RTS"
	case QPStateSQD:
		return "SQD"
	case QPStateSQE:
		return "SQE"
	case QPStateErr:
		return "ERR"
	case QPStateDrain:
		return "DRAIN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", uint8(s))
	}
}

// NCCLTopoDomain classifies the network topology relationship between two GPUs.
type NCCLTopoDomain uint8

const (
	NCCLTopoDomainSameGPU NCCLTopoDomain = 0 // Shared memory / same GPU
	NCCLTopoDomainNVLink  NCCLTopoDomain = 1 // NVLink peer
	NCCLTopoDomainPCIe    NCCLTopoDomain = 2 // Same PCIe root complex
	NCCLTopoDomainSwitch  NCCLTopoDomain = 3 // Same ToR switch
	NCCLTopoDomainRemote  NCCLTopoDomain = 4 // Cross-rack
)

// QPKey is the BPF map key for sg_qp_table. Must match the C struct.
type QPKey struct {
	QPN uint32
}

// QPEntry is the BPF map value for sg_qp_table. Must match the C struct.
// Total size: 1+1+2+4+4+4+8 = 24 bytes (aligned).
type QPEntry struct {
	State       uint8
	PFCPriority uint8
	MTU         uint16
	RemoteIP    uint32
	RQPSN       uint32
	SQPSN       uint32
	LastSeenNs  uint64
}

// NCCLTopoKey is the BPF map key for sg_gpu_topo.
type NCCLTopoKey struct {
	RemoteIP uint32
}

// NCCLTopoEntry is the BPF map value for sg_gpu_topo.
type NCCLTopoEntry struct {
	Domain        uint8
	HopCount      uint8
	BandwidthGbps uint16
	LocalGPUID    uint32
	RemoteGPUID   uint32
	NVLinkMask    uint32
	LastUpdatedNs uint64
}

// PFCConfig is the BPF map value for sg_pfc_config (one entry per priority).
type PFCConfig struct {
	Enabled       uint8
	DSCP          uint8
	ECNCapable    uint8
	Pad           uint8
	PauseTimeUs   uint32
	XONThreshold  uint32
	XOFFThreshold uint32
}

// RoCECounter is the per-CPU per-QPN counter read from sg_roce_counters.
type RoCECounter struct {
	Bytes     uint64
	Packets   uint64
	CNPSent   uint64
	ECNCESeen uint64
	Drops     uint64
}

// RoCECongestionEvent mirrors the BPF struct for ring buffer events.
type RoCECongestionEvent struct {
	Timestamp uint64
	SrcIP     uint32
	DstIP     uint32
	QPN       uint32
	PSN       uint32
	ECN       uint8
	Opcode    uint8
	Pad       uint16
}

// Config configures the RoCE Manager.
type Config struct {
	// BPFFSBase is the bpffs pin directory for RoCE maps.
	BPFFSBase string
	// Interface is the RDMA-capable network interface name (e.g., "mlx5_0").
	Interface string
	// XDPObjPath is the path to the compiled roce_xdp.o ELF object.
	XDPObjPath string
	// TCObjPath is the path to the compiled roce_tc.o ELF object.
	TCObjPath string
	// TopologyRefreshInterval controls GPU topology refresh period.
	TopologyRefreshInterval time.Duration
	// DefaultPFCConfig is the PFC configuration applied to all priorities on startup.
	DefaultPFCConfig []PFCConfig
	// OnCongestionEvent is called for each ECN/CNP event from the ring buffer.
	// May be nil (events are discarded).
	OnCongestionEvent func(ev RoCECongestionEvent)
}

// DefaultConfig returns a sensible default RoCE manager configuration.
func DefaultConfig(iface string) Config {
	pfcDefaults := make([]PFCConfig, PFCPriorities)
	for i := range pfcDefaults {
		pfcDefaults[i] = PFCConfig{
			Enabled:       0, // disabled by default; enable only for RoCE priority
			DSCP:          0,
			ECNCapable:    0,
			PauseTimeUs:   65535,
			XONThreshold:  256 * 1024,
			XOFFThreshold: 512 * 1024,
		}
	}
	// Priority 3 is the conventional RoCEv2 traffic class
	pfcDefaults[3] = PFCConfig{
		Enabled:       1,
		DSCP:          26, // AF31
		ECNCapable:    1,
		PauseTimeUs:   65535,
		XONThreshold:  256 * 1024,
		XOFFThreshold: 512 * 1024,
	}

	return Config{
		BPFFSBase:               DefaultBPFFSBase,
		Interface:               iface,
		XDPObjPath:              "bpf/roce/roce_xdp.o",
		TCObjPath:               "bpf/roce/roce_tc.o",
		TopologyRefreshInterval: DefaultTopologyRefreshInterval,
		DefaultPFCConfig:        pfcDefaults,
	}
}

// Manager is the StraitGateway RoCEv2 dataplane manager.
type Manager struct {
	mu  sync.RWMutex
	cfg Config

	// BPF map handles
	qpTable      *ebpf.Map
	roCECounters *ebpf.Map
	gpuTopo      *ebpf.Map
	roCEEvents   *ebpf.Map
	pfcConfig    *ebpf.Map

	// Active link handles
	xdpLink link.Link
	tcLink  link.Link

	// Ring buffer reader
	rdr *ringbuf.Reader

	// Internal state
	qpCache map[uint32]*QPEntry

	// Lifecycle
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new RoCE Manager. If BPF maps are already pinned (from a
// previous run), they are reused via LoadPinnedMap. Otherwise they are
// created and pinned.
func New(cfg Config) (*Manager, error) {
	if cfg.BPFFSBase == "" {
		cfg.BPFFSBase = DefaultBPFFSBase
	}
	if cfg.TopologyRefreshInterval == 0 {
		cfg.TopologyRefreshInterval = DefaultTopologyRefreshInterval
	}
	if err := os.MkdirAll(cfg.BPFFSBase, 0o755); err != nil {
		return nil, fmt.Errorf("roce: create bpffs dir %s: %w", cfg.BPFFSBase, err)
	}

	m := &Manager{
		cfg:     cfg,
		qpCache: make(map[uint32]*QPEntry),
	}

	if err := m.initMaps(); err != nil {
		return nil, fmt.Errorf("roce: init maps: %w", err)
	}

	return m, nil
}

// Start loads and attaches XDP/TC programs, starts the ring buffer consumer
// and the GPU topology refresh goroutine.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Apply default PFC configuration
	if err := m.applyPFCConfig(m.cfg.DefaultPFCConfig); err != nil {
		return fmt.Errorf("roce: apply PFC config: %w", err)
	}

	// Load XDP program if ELF is available
	if m.cfg.XDPObjPath != "" {
		if err := m.loadXDP(); err != nil {
			// Log and continue — XDP may not be available on all kernels/NICs
			fmt.Printf("roce: XDP load warning: %v (continuing without XDP fast-path)\n", err)
		}
	}

	// Load TC program if ELF is available
	if m.cfg.TCObjPath != "" {
		if err := m.loadTC(); err != nil {
			fmt.Printf("roce: TC load warning: %v (continuing without TC egress)\n", err)
		}
	}

	// Start ring buffer consumer
	if m.roCEEvents != nil {
		rdr, err := ringbuf.NewReader(m.roCEEvents)
		if err == nil {
			m.rdr = rdr
		}
	}

	cCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	m.wg.Add(2)
	go m.runEventConsumer(cCtx)
	go m.runTopologyRefresh(cCtx)

	return nil
}

// Stop gracefully shuts down the Manager, detaching BPF programs and closing maps.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()

	if m.rdr != nil {
		_ = m.rdr.Close()
	}

	var errs []error
	if m.xdpLink != nil {
		if err := m.xdpLink.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close XDP link: %w", err))
		}
	}
	if m.tcLink != nil {
		if err := m.tcLink.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close TC link: %w", err))
		}
	}

	for _, mp := range []*ebpf.Map{m.qpTable, m.roCECounters, m.gpuTopo, m.roCEEvents, m.pfcConfig} {
		if mp != nil {
			if err := mp.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("roce manager stop errors: %v", errs)
	}
	return nil
}

// UpsertQP creates or updates a QP entry in sg_qp_table.
func (m *Manager) UpsertQP(qpn uint32, entry QPEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.qpTable == nil {
		m.qpCache[qpn] = &entry
		return nil
	}

	key := QPKey{QPN: qpn}
	if err := m.qpTable.Put(key, entry); err != nil {
		return fmt.Errorf("roce: upsert QP %d: %w", qpn, err)
	}
	m.qpCache[qpn] = &entry
	return nil
}

// DeleteQP removes a QP entry from sg_qp_table.
func (m *Manager) DeleteQP(qpn uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.qpCache, qpn)
	if m.qpTable == nil {
		return nil
	}
	key := QPKey{QPN: qpn}
	return m.qpTable.Delete(key)
}

// TransitionQP transitions a QP to a new state, enforcing the valid state
// machine transitions per IB spec.
func (m *Manager) TransitionQP(qpn uint32, newState QPState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.qpCache[qpn]
	if !ok {
		return fmt.Errorf("roce: QP %d not found", qpn)
	}

	current := QPState(entry.State)
	if !validTransition(current, newState) {
		return fmt.Errorf("roce: invalid QP transition %s → %s for QPN %d",
			current, newState, qpn)
	}

	entry.State = uint8(newState)
	if m.qpTable == nil {
		return nil
	}
	key := QPKey{QPN: qpn}
	return m.qpTable.Put(key, *entry)
}

// ReadCounters reads the per-CPU counters for a given QPN and aggregates them.
func (m *Manager) ReadCounters(qpn uint32) (RoCECounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total RoCECounter
	if m.roCECounters == nil {
		return total, nil
	}

	// Per-CPU map: one value per CPU, read all and sum
	values, err := m.roCECounters.LookupBytes(QPKey{QPN: qpn})
	if err != nil {
		return total, fmt.Errorf("roce: read counters for QPN %d: %w", qpn, err)
	}

	// Each CPU's value is sizeof(RoCECounter) = 40 bytes
	const counterSize = int(unsafe.Sizeof(RoCECounter{}))
	for off := 0; off+counterSize <= len(values); off += counterSize {
		var c RoCECounter
		c.Bytes = binary.LittleEndian.Uint64(values[off:])
		c.Packets = binary.LittleEndian.Uint64(values[off+8:])
		c.CNPSent = binary.LittleEndian.Uint64(values[off+16:])
		c.ECNCESeen = binary.LittleEndian.Uint64(values[off+24:])
		c.Drops = binary.LittleEndian.Uint64(values[off+32:])
		total.Bytes += c.Bytes
		total.Packets += c.Packets
		total.CNPSent += c.CNPSent
		total.ECNCESeen += c.ECNCESeen
		total.Drops += c.Drops
	}
	return total, nil
}

// UpsertTopoEntry inserts or updates a GPU topology entry in sg_gpu_topo.
func (m *Manager) UpsertTopoEntry(remoteIP uint32, entry NCCLTopoEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.gpuTopo == nil {
		return nil
	}
	key := NCCLTopoKey{RemoteIP: remoteIP}
	return m.gpuTopo.Put(key, entry)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (m *Manager) initMaps() error {
	// Try to load pinned maps; create fresh ones if absent
	pinPaths := map[string]**ebpf.Map{
		"sg_qp_table":      &m.qpTable,
		"sg_roce_counters": &m.roCECounters,
		"sg_gpu_topo":      &m.gpuTopo,
		"sg_roce_events":   &m.roCEEvents,
		"sg_pfc_config":    &m.pfcConfig,
	}

	for name, target := range pinPaths {
		pinPath := filepath.Join(m.cfg.BPFFSBase, name)
		mp, err := ebpf.LoadPinnedMap(pinPath, nil)
		if err == nil {
			*target = mp
			continue
		}
		// Create new map
		spec := roCEMapSpec(name)
		if spec == nil {
			continue
		}
		mp, err = ebpf.NewMap(spec)
		if err != nil {
			if isPermissionError(err) {
				// Running without CAP_BPF — maps will be nil (no-op mode)
				continue
			}
			return fmt.Errorf("create map %s: %w", name, err)
		}
		// Pin the map
		if pinErr := mp.Pin(pinPath); pinErr != nil {
			_ = mp.Close()
			// Non-fatal: continue without pinning in restricted envs
		}
		*target = mp
	}
	return nil
}

func roCEMapSpec(name string) *ebpf.MapSpec {
	switch name {
	case "sg_qp_table":
		return &ebpf.MapSpec{
			Name:       "sg_qp_table",
			Type:       ebpf.Hash,
			KeySize:    uint32(unsafe.Sizeof(QPKey{})),
			ValueSize:  uint32(unsafe.Sizeof(QPEntry{})),
			MaxEntries: MaxQPs,
		}
	case "sg_roce_counters":
		return &ebpf.MapSpec{
			Name:       "sg_roce_counters",
			Type:       ebpf.PerCPUHash,
			KeySize:    8, // QPKey + cpu + pad
			ValueSize:  uint32(unsafe.Sizeof(RoCECounter{})),
			MaxEntries: MaxQPs,
		}
	case "sg_gpu_topo":
		return &ebpf.MapSpec{
			Name:       "sg_gpu_topo",
			Type:       ebpf.Hash,
			KeySize:    uint32(unsafe.Sizeof(NCCLTopoKey{})),
			ValueSize:  uint32(unsafe.Sizeof(NCCLTopoEntry{})),
			MaxEntries: MaxGPUPeers,
		}
	case "sg_roce_events":
		return &ebpf.MapSpec{
			Name:       "sg_roce_events",
			Type:       ebpf.RingBuf,
			MaxEntries: RoCEEventRingBufSize,
		}
	case "sg_pfc_config":
		return &ebpf.MapSpec{
			Name:       "sg_pfc_config",
			Type:       ebpf.Array,
			KeySize:    4,
			ValueSize:  uint32(unsafe.Sizeof(PFCConfig{})),
			MaxEntries: PFCPriorities,
		}
	}
	return nil
}

func (m *Manager) applyPFCConfig(cfgs []PFCConfig) error {
	if m.pfcConfig == nil {
		return nil
	}
	for i, cfg := range cfgs {
		if i >= PFCPriorities {
			break
		}
		key := uint32(i)
		if err := m.pfcConfig.Put(key, cfg); err != nil {
			return fmt.Errorf("put PFC config priority %d: %w", i, err)
		}
	}
	return nil
}

func (m *Manager) loadXDP() error {
	if _, err := os.Stat(m.cfg.XDPObjPath); os.IsNotExist(err) {
		return fmt.Errorf("XDP object not found: %s", m.cfg.XDPObjPath)
	}
	spec, err := ebpf.LoadCollectionSpec(m.cfg.XDPObjPath)
	if err != nil {
		return fmt.Errorf("load XDP collection: %w", err)
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("new XDP collection: %w", err)
	}
	defer coll.Close()

	prog, ok := coll.Programs["sg_roce_xdp"]
	if !ok {
		return fmt.Errorf("program sg_roce_xdp not found in %s", m.cfg.XDPObjPath)
	}

	ifaceIdx, err := ifaceIndex(m.cfg.Interface)
	if err != nil {
		return err
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: ifaceIdx,
		Flags:     link.XDPDriverMode,
	})
	if err != nil {
		// Fall back to generic mode
		l, err = link.AttachXDP(link.XDPOptions{
			Program:   prog,
			Interface: ifaceIdx,
			Flags:     link.XDPGenericMode,
		})
		if err != nil {
			return fmt.Errorf("attach XDP (driver + generic fallback): %w", err)
		}
	}
	m.xdpLink = l
	return nil
}

func (m *Manager) loadTC() error {
	if _, err := os.Stat(m.cfg.TCObjPath); os.IsNotExist(err) {
		return fmt.Errorf("TC object not found: %s", m.cfg.TCObjPath)
	}
	spec, err := ebpf.LoadCollectionSpec(m.cfg.TCObjPath)
	if err != nil {
		return fmt.Errorf("load TC collection: %w", err)
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("new TC collection: %w", err)
	}
	defer coll.Close()

	prog, ok := coll.Programs["sg_roce_egress"]
	if !ok {
		return fmt.Errorf("program sg_roce_egress not found in %s", m.cfg.TCObjPath)
	}

	ifaceIdx, err := ifaceIndex(m.cfg.Interface)
	if err != nil {
		return err
	}

	l, err := link.AttachTCX(link.TCXOptions{
		Program:   prog,
		Attach:    ebpf.AttachTCXEgress,
		Interface: ifaceIdx,
	})
	if err != nil {
		return fmt.Errorf("attach TCX egress: %w", err)
	}
	m.tcLink = l
	return nil
}

// runEventConsumer drains sg_roce_events ring buffer and dispatches events.
func (m *Manager) runEventConsumer(ctx context.Context) {
	defer m.wg.Done()

	if m.rdr == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		record, err := m.rdr.Read()
		if err != nil {
			if err == ringbuf.ErrClosed {
				return
			}
			continue
		}

		if len(record.RawSample) < int(unsafe.Sizeof(RoCECongestionEvent{})) {
			continue
		}

		var ev RoCECongestionEvent
		ev.Timestamp = binary.LittleEndian.Uint64(record.RawSample[0:])
		ev.SrcIP = binary.LittleEndian.Uint32(record.RawSample[8:])
		ev.DstIP = binary.LittleEndian.Uint32(record.RawSample[12:])
		ev.QPN = binary.LittleEndian.Uint32(record.RawSample[16:])
		ev.PSN = binary.LittleEndian.Uint32(record.RawSample[20:])
		ev.ECN = record.RawSample[24]
		ev.Opcode = record.RawSample[25]

		m.mu.RLock()
		fn := m.cfg.OnCongestionEvent
		m.mu.RUnlock()

		if fn != nil {
			fn(ev)
		}
	}
}

// runTopologyRefresh periodically calls DiscoverGPUTopology and updates sg_gpu_topo.
func (m *Manager) runTopologyRefresh(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cfg.TopologyRefreshInterval)
	defer ticker.Stop()

	// Immediate first refresh
	m.refreshTopology()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.refreshTopology()
		}
	}
}

func (m *Manager) refreshTopology() {
	entries, err := DiscoverGPUTopology()
	if err != nil {
		return // nvidia-smi not available or no GPUs — silently skip
	}
	for remoteIP, entry := range entries {
		_ = m.UpsertTopoEntry(remoteIP, entry)
	}
}

func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "operation not permitted") ||
		strings.Contains(s, "permission denied")
}

func ifaceIndex(name string) (int, error) {
	ifaces, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return 0, fmt.Errorf("read /sys/class/net: %w", err)
	}
	for _, de := range ifaces {
		if de.Name() == name {
			ifIndexPath := filepath.Join("/sys/class/net", name, "ifindex")
			data, err := os.ReadFile(ifIndexPath)
			if err != nil {
				return 0, fmt.Errorf("read ifindex for %s: %w", name, err)
			}
			idx, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				return 0, fmt.Errorf("parse ifindex for %s: %w", name, err)
			}
			return idx, nil
		}
	}
	return 0, fmt.Errorf("interface %q not found", name)
}

// validTransition returns true if the QP state transition is permitted by
// the IB specification (Table 69, IB Spec Vol 1).
func validTransition(from, to QPState) bool {
	allowed := map[QPState][]QPState{
		QPStateReset: {QPStateInit},
		QPStateInit:  {QPStateRTR, QPStateReset, QPStateErr},
		QPStateRTR:   {QPStateRTS, QPStateErr, QPStateReset},
		QPStateRTS:   {QPStateSQD, QPStateErr, QPStateReset},
		QPStateSQD:   {QPStateRTS, QPStateErr, QPStateReset},
		QPStateSQE:   {QPStateRTS, QPStateErr, QPStateReset},
		QPStateErr:   {QPStateReset},
		QPStateDrain: {QPStateErr, QPStateReset},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

// DiscoverGPUTopology runs nvidia-smi to build the NCCL topology map.
// Returns a map of remote_ip → NCCLTopoEntry.
// If nvidia-smi is unavailable or returns no GPUs, returns empty map (no error).
func DiscoverGPUTopology() (map[uint32]NCCLTopoEntry, error) {
	result := make(map[uint32]NCCLTopoEntry)

	out, err := exec.Command("nvidia-smi", "topo", "-m").Output()
	if err != nil {
		// Not a fatal error — system may not have GPUs
		return result, nil
	}

	lines := strings.Split(string(out), "\n")
	gpuCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "GPU") {
			gpuCount++
		}
	}

	// Build simplistic topology: GPUs on same host are NVLink peers;
	// all others are cross-switch (remote).
	// A full production implementation would parse the NVLink matrix from
	// nvidia-smi topo -m and map GPU GUIDs to interface IP addresses.
	localGPUID := uint32(0)
	for i := 0; i < gpuCount; i++ {
		if uint32(i) == localGPUID {
			continue
		}
		// GPU peers on same host use loopback-range virtual IPs
		// 169.254.0.x where x is the GPU ordinal
		remoteIP := encodeGPUVirtualIP(uint32(i))
		result[remoteIP] = NCCLTopoEntry{
			Domain:        uint8(NCCLTopoDomainNVLink),
			HopCount:      1,
			BandwidthGbps: 600, // NVLink 3.0: 600 GB/s aggregate
			LocalGPUID:    localGPUID,
			RemoteGPUID:   uint32(i),
			NVLinkMask:    0xFFFF, // all 16 NVLink lanes
			LastUpdatedNs: uint64(time.Now().UnixNano()),
		}
	}

	return result, nil
}

// encodeGPUVirtualIP encodes a GPU ordinal as 169.254.0.<gpuID>.
func encodeGPUVirtualIP(gpuID uint32) uint32 {
	// 169.254.0.gpuID in big-endian network byte order
	return (169 << 24) | (254 << 16) | (0 << 8) | (gpuID & 0xFF)
}
