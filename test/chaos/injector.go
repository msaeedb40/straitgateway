// Package chaos provides a userspace eBPF chaos injector for StraitGateway.
//
// The injector uses BPF_PROG_TEST_RUN (prog_test_run) to validate program
// behaviour under adversarial conditions, and writes chaos configurations
// into special BPF control maps that are read by eBPF programs at runtime
// to simulate:
//
//   - ENOMEM — memory allocation failures in BPF helpers
//   - Packet loss burst — drops N consecutive packets matching a selector
//   - Latency jitter — adds synthetic RTT jitter via tc netem (userspace)
//   - BFD failover — flaps a route/link to trigger BFD session failures
//
// Usage:
//
//	injector := chaos.New(chaos.Config{...})
//	injector.InjectENOMEM(ctx, 500*time.Millisecond)
//	injector.InjectDropBurst(ctx, 32, chaos.DropBurstConfig{...})
package chaos

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
)

// FaultType enumerates the supported chaos fault modes.
type FaultType uint32

const (
	FaultNone      FaultType = 0
	FaultENOMEM    FaultType = 1
	FaultDropBurst FaultType = 2
	FaultJitter    FaultType = 3
	FaultBFDFlap   FaultType = 4
)

// chaosMapKey is the BPF map key used for per-fault-type configuration.
type chaosMapKey uint32

// chaosMapValue encodes the fault parameters stored in the BPF control map.
// Layout matches the C struct in bpf/chaos/chaos_ctrl.h (if present).
//
//	struct chaos_ctrl {
//	    __u32 fault_type;
//	    __u32 drop_count;
//	    __u32 drop_probability; // out of 100
//	    __u32 jitter_ms;
//	    __u64 expiry_ns;        // ktime_get_ns() + duration
//	};
type chaosMapValue struct {
	FaultType       uint32
	DropCount       uint32
	DropProbability uint32
	JitterMs        uint32
	ExpiryNs        uint64
}

// DropBurstConfig configures the packet drop burst fault.
type DropBurstConfig struct {
	// BurstSize is the number of consecutive packets to drop.
	BurstSize uint32
	// Probability is the probability (0-100) that the burst is triggered per packet.
	Probability uint32
}

// JitterConfig configures the network latency jitter fault.
type JitterConfig struct {
	// Interface is the netlink interface name to apply tc-netem on.
	Interface string
	// BaseDelayMs is the base RTT delay in milliseconds.
	BaseDelayMs uint32
	// JitterMs is the jitter range in milliseconds.
	JitterMs uint32
}

// BFDFlapConfig configures the BFD failover fault.
type BFDFlapConfig struct {
	// PeerIP is the BFD peer IP address.
	PeerIP net.IP
	// Interface is the interface the BFD session is established on.
	Interface string
	// FlapCount is the number of times to flap the link.
	FlapCount int
	// FlapIntervalMs is the interval between flaps.
	FlapIntervalMs int
}

// Config is the Injector configuration.
type Config struct {
	// ChaosMapPath is the pinned bpffs path to the chaos control map.
	// If empty, the injector creates an in-process ephemeral map.
	ChaosMapPath string
	// BPFObjPath is the path to the test eBPF program ELF (for prog_test_run).
	BPFObjPath string
}

// Injector manages fault injection into the StraitGateway dataplane.
type Injector struct {
	mu       sync.Mutex
	cfg      Config
	ctrlMap  *ebpf.Map
	cancelFn context.CancelFunc
}

// New creates an Injector.  If cfg.ChaosMapPath is non-empty, it tries to
// load the pinned map; otherwise it allocates a process-local BPF hash map.
func New(cfg Config) (*Injector, error) {
	var ctrlMap *ebpf.Map

	if cfg.ChaosMapPath != "" {
		var err error
		ctrlMap, err = ebpf.LoadPinnedMap(cfg.ChaosMapPath, nil)
		if err != nil {
			return nil, fmt.Errorf("chaos: load pinned map %s: %w", cfg.ChaosMapPath, err)
		}
	} else {
		var err error
		ctrlMap, err = ebpf.NewMap(&ebpf.MapSpec{
			Name:       "sg_chaos_ctrl",
			Type:       ebpf.Hash,
			KeySize:    uint32(unsafe.Sizeof(chaosMapKey(0))),
			ValueSize:  uint32(unsafe.Sizeof(chaosMapValue{})),
			MaxEntries: 16,
		})
		if err != nil {
			// On unsupported kernels (CI without BPF) fall back to a no-op map
			if isUnsupportedKernel(err) {
				ctrlMap = nil
			} else {
				return nil, fmt.Errorf("chaos: create control map: %w", err)
			}
		}
	}

	return &Injector{cfg: cfg, ctrlMap: ctrlMap}, nil
}

// InjectENOMEM simulates ENOMEM helper failures for the given duration.
// Writes a chaos control record with FaultENOMEM into the BPF map, which
// causes any BPF program that reads sg_chaos_ctrl to return early with ENOMEM.
func (inj *Injector) InjectENOMEM(ctx context.Context, duration time.Duration) error {
	inj.mu.Lock()
	defer inj.mu.Unlock()

	val := chaosMapValue{
		FaultType: uint32(FaultENOMEM),
		ExpiryNs:  uint64(time.Now().Add(duration).UnixNano()),
	}
	if err := inj.writeCtrlMap(FaultENOMEM, val); err != nil {
		return fmt.Errorf("InjectENOMEM: %w", err)
	}

	go inj.expireAfter(ctx, FaultENOMEM, duration)
	return nil
}

// InjectDropBurst triggers a packet drop burst on the next BurstSize packets.
// The eBPF program uses a per-CPU counter in the chaos map to count drops.
func (inj *Injector) InjectDropBurst(ctx context.Context, duration time.Duration, cfg DropBurstConfig) error {
	inj.mu.Lock()
	defer inj.mu.Unlock()

	val := chaosMapValue{
		FaultType:       uint32(FaultDropBurst),
		DropCount:       cfg.BurstSize,
		DropProbability: cfg.Probability,
		ExpiryNs:        uint64(time.Now().Add(duration).UnixNano()),
	}
	if err := inj.writeCtrlMap(FaultDropBurst, val); err != nil {
		return fmt.Errorf("InjectDropBurst: %w", err)
	}

	go inj.expireAfter(ctx, FaultDropBurst, duration)
	return nil
}

// InjectJitter applies tc-netem latency jitter to the given network interface.
// Uses `tc qdisc add dev <iface> root netem delay <base>ms <jitter>ms`.
func (inj *Injector) InjectJitter(ctx context.Context, duration time.Duration, cfg JitterConfig) error {
	inj.mu.Lock()
	defer inj.mu.Unlock()

	if cfg.Interface == "" {
		return fmt.Errorf("InjectJitter: Interface must be set")
	}

	// Add netem qdisc
	if err := runCmd("tc", "qdisc", "add", "dev", cfg.Interface,
		"root", "netem",
		"delay", fmt.Sprintf("%dms", cfg.BaseDelayMs),
		fmt.Sprintf("%dms", cfg.JitterMs)); err != nil {
		// Try replace if already exists
		_ = runCmd("tc", "qdisc", "change", "dev", cfg.Interface,
			"root", "netem",
			"delay", fmt.Sprintf("%dms", cfg.BaseDelayMs),
			fmt.Sprintf("%dms", cfg.JitterMs))
	}

	// Also record in BPF map for eBPF-aware components
	val := chaosMapValue{
		FaultType: uint32(FaultJitter),
		JitterMs:  cfg.JitterMs,
		ExpiryNs:  uint64(time.Now().Add(duration).UnixNano()),
	}
	_ = inj.writeCtrlMap(FaultJitter, val)

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(duration):
		}
		// Remove netem qdisc
		_ = runCmd("tc", "qdisc", "del", "dev", cfg.Interface, "root")
		_ = inj.clearCtrlMap(FaultJitter)
	}()

	return nil
}

// InjectBFDFlap simulates BFD session failure by flapping the given interface.
// Each flap: link down (FlapIntervalMs/2) → link up (FlapIntervalMs/2).
func (inj *Injector) InjectBFDFlap(ctx context.Context, cfg BFDFlapConfig) error {
	if cfg.Interface == "" {
		return fmt.Errorf("InjectBFDFlap: Interface must be set")
	}

	go func() {
		for i := 0; i < cfg.FlapCount; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Link down
			_ = runCmd("ip", "link", "set", "dev", cfg.Interface, "down")
			time.Sleep(time.Duration(cfg.FlapIntervalMs/2) * time.Millisecond)

			// Link up
			_ = runCmd("ip", "link", "set", "dev", cfg.Interface, "up")
			time.Sleep(time.Duration(cfg.FlapIntervalMs/2) * time.Millisecond)
		}
	}()

	return nil
}

// RunProgTestRun executes a compiled eBPF program against a synthetic packet
// using BPF_PROG_TEST_RUN. It returns the program return code and data output.
//
// This is used to validate eBPF program behaviour under each fault type
// without attaching the program to a live interface.
func (inj *Injector) RunProgTestRun(progSection string, inputPkt []byte) (uint32, []byte, error) {
	if inj.cfg.BPFObjPath == "" {
		return 0, nil, fmt.Errorf("RunProgTestRun: BPFObjPath not configured")
	}

	spec, err := ebpf.LoadCollectionSpec(inj.cfg.BPFObjPath)
	if err != nil {
		return 0, nil, fmt.Errorf("RunProgTestRun: load spec %s: %w", inj.cfg.BPFObjPath, err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return 0, nil, fmt.Errorf("RunProgTestRun: new collection: %w", err)
	}
	defer coll.Close()

	prog, ok := coll.Programs[progSection]
	if !ok {
		return 0, nil, fmt.Errorf("RunProgTestRun: program section %q not found", progSection)
	}

	ret, out, err := prog.Test(inputPkt)
	if err != nil {
		return 0, nil, fmt.Errorf("RunProgTestRun: prog test run: %w", err)
	}
	return ret, out, nil
}

// BuildMinimalPassProg creates an ephemeral eBPF program that always returns
// XDP_PASS (2) — useful as a baseline for chaos regression tests that need a
// real BPF prog but don't have the full StraitGateway ELF available.
func BuildMinimalPassProg() (*ebpf.Program, error) {
	insns := asm.Instructions{
		asm.Mov.Imm(asm.R0, 2), // XDP_PASS
		asm.Return(),
	}
	spec := &ebpf.ProgramSpec{
		Name:         "chaos_pass",
		Type:         ebpf.XDP,
		License:      "GPL",
		Instructions: insns,
	}
	return ebpf.NewProgram(spec)
}

// ClearAll removes all chaos fault records from the control map.
func (inj *Injector) ClearAll() error {
	inj.mu.Lock()
	defer inj.mu.Unlock()

	for _, ft := range []FaultType{FaultENOMEM, FaultDropBurst, FaultJitter, FaultBFDFlap} {
		_ = inj.clearCtrlMap(ft)
	}
	return nil
}

// Close releases resources held by the Injector.
func (inj *Injector) Close() error {
	inj.mu.Lock()
	defer inj.mu.Unlock()
	if inj.ctrlMap != nil {
		return inj.ctrlMap.Close()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (inj *Injector) writeCtrlMap(ft FaultType, val chaosMapValue) error {
	if inj.ctrlMap == nil {
		return nil // no-op in CI without BPF
	}
	key := chaosMapKey(ft)
	return inj.ctrlMap.Put(key, val)
}

func (inj *Injector) clearCtrlMap(ft FaultType) error {
	if inj.ctrlMap == nil {
		return nil
	}
	key := chaosMapKey(ft)
	return inj.ctrlMap.Delete(key)
}

func (inj *Injector) expireAfter(ctx context.Context, ft FaultType, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
	inj.mu.Lock()
	defer inj.mu.Unlock()
	_ = inj.clearCtrlMap(ft)
}

func runCmd(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isUnsupportedKernel(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return contains(s, "operation not permitted") ||
		contains(s, "not supported") ||
		contains(s, "permission denied")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// syntheticUDPPacket builds a minimal raw Ethernet+IP+UDP packet for prog_test_run.
func syntheticUDPPacket(srcIP, dstIP net.IP, srcPort, dstPort uint16) []byte {
	pkt := make([]byte, 14+20+8+8) // Eth+IPv4+UDP+payload
	// Ethernet: dst=ff:ff:ff:ff:ff:ff src=de:ad:be:ef:00:01 type=0x0800
	copy(pkt[0:], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	copy(pkt[6:], []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x01})
	binary.BigEndian.PutUint16(pkt[12:], 0x0800)

	// IPv4
	pkt[14] = 0x45 // version+IHL
	pkt[15] = 0x00 // DSCP
	binary.BigEndian.PutUint16(pkt[16:], uint16(len(pkt)-14))
	binary.BigEndian.PutUint16(pkt[18:], uint16(rand.Intn(65535)))
	pkt[22] = 0x40 // DF
	pkt[23] = 0x00
	pkt[24] = 64                 // TTL
	pkt[25] = 17                 // UDP
	copy(pkt[26:], []byte{0, 0}) // checksum (0 = let kernel compute)
	copy(pkt[28:], srcIP.To4())
	copy(pkt[32:], dstIP.To4())

	// UDP
	binary.BigEndian.PutUint16(pkt[34:], srcPort)
	binary.BigEndian.PutUint16(pkt[36:], dstPort)
	binary.BigEndian.PutUint16(pkt[38:], 8+8)
	binary.BigEndian.PutUint16(pkt[40:], 0) // checksum

	// payload
	copy(pkt[42:], []byte("sgchaos!"))
	return pkt
}

// SyntheticUDPPacket is an exported convenience for test files.
func SyntheticUDPPacket(srcIP, dstIP net.IP, srcPort, dstPort uint16) []byte {
	return syntheticUDPPacket(srcIP, dstIP, srcPort, dstPort)
}
