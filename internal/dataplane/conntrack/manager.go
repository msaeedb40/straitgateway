package conntrack

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// FlowDirection indicates forward or reverse flow
type FlowDirection int

const (
	DirForward FlowDirection = iota
	DirReverse
)

// FlowState tracks high-level TCP/UDP connection state
type FlowState struct {
	SrcIP      net.IP
	DstIP      net.IP
	SrcPort    uint16
	DstPort    uint16
	Protocol   uint8
	Packets    uint64
	Bytes      uint64
	LastSeen   time.Time
	RevSrcIP   net.IP
	RevDstIP   net.IP
	RevSrcPort uint16
	RevDstPort uint16
}

// Manager coordinates 5-tuple conntrack lifecycle with BPF maps
type Manager struct {
	mu     sync.RWMutex
	mapMgr *maps.MapManager
}

// NewManager creates a conntrack manager
func NewManager(mapMgr *maps.MapManager) *Manager {
	return &Manager{
		mapMgr: mapMgr,
	}
}

// ipToU32 converts net.IP to big-endian uint32
func ipToU32(ip net.IP) uint32 {
	v4 := ip.To4()
	if v4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(v4)
}

// u32ToIP converts uint32 to net.IP
func u32ToIP(val uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, val)
	return net.IP(b)
}

// BuildFlowKey constructs a 5-tuple flow key
func BuildFlowKey(srcIP, dstIP net.IP, srcPort, dstPort uint16, proto uint8) types.FlowKey {
	return types.FlowKey{
		SrcIP:   ipToU32(srcIP),
		DstIP:   ipToU32(dstIP),
		SrcPort: srcPort,
		DstPort: dstPort,
		Proto:   proto,
	}
}

// TrackFlow registers or updates a bidirectional conntrack entry
func (m *Manager) TrackFlow(state FlowState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mapMgr == nil {
		return nil
	}

	key := BuildFlowKey(state.SrcIP, state.DstIP, state.SrcPort, state.DstPort, state.Protocol)
	entry := types.FlowEntry{
		Packets:    state.Packets,
		Bytes:      state.Bytes,
		LastSeen:   uint64(time.Now().UnixNano()),
		Flags:      1, // ESTABLISHED
		RevSrcIP:   ipToU32(state.RevSrcIP),
		RevDstIP:   ipToU32(state.RevDstIP),
		RevSrcPort: state.RevSrcPort,
		RevDstPort: state.RevDstPort,
	}

	return m.mapMgr.SetConntrack(key, entry)
}

// LookupFlow queries conntrack state for a 5-tuple
func (m *Manager) LookupFlow(srcIP, dstIP net.IP, srcPort, dstPort uint16, proto uint8) (*FlowState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mapMgr == nil {
		return nil, fmt.Errorf("map manager not initialized")
	}

	key := BuildFlowKey(srcIP, dstIP, srcPort, dstPort, proto)
	entry, err := m.mapMgr.LookupConntrack(key)
	if err != nil {
		return nil, err
	}

	return &FlowState{
		SrcIP:      srcIP,
		DstIP:      dstIP,
		SrcPort:    srcPort,
		DstPort:    dstPort,
		Protocol:   proto,
		Packets:    entry.Packets,
		Bytes:      entry.Bytes,
		LastSeen:   time.Unix(0, int64(entry.LastSeen)),
		RevSrcIP:   u32ToIP(entry.RevSrcIP),
		RevDstIP:   u32ToIP(entry.RevDstIP),
		RevSrcPort: entry.RevSrcPort,
		RevDstPort: entry.RevDstPort,
	}, nil
}

// PurgeFlow removes a conntrack entry from the dataplane
func (m *Manager) PurgeFlow(srcIP, dstIP net.IP, srcPort, dstPort uint16, proto uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mapMgr == nil {
		return nil
	}

	key := BuildFlowKey(srcIP, dstIP, srcPort, dstPort, proto)
	return m.mapMgr.DeleteConntrack(key)
}
