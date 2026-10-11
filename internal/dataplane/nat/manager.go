package nat

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// Translation defines an L4 NAT translation rule
type Translation struct {
	OrigIP    net.IP
	OrigPort  uint16
	TransIP   net.IP
	TransPort uint16
	Protocol  uint8
}

// Manager coordinates SNAT and DNAT map entries with the eBPF dataplane
type Manager struct {
	mu     sync.RWMutex
	mapMgr *maps.MapManager
}

// NewManager creates a NAT Manager
func NewManager(mapMgr *maps.MapManager) *Manager {
	return &Manager{
		mapMgr: mapMgr,
	}
}

func ipToU32(ip net.IP) uint32 {
	v4 := ip.To4()
	if v4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(v4)
}

func u32ToIP(val uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, val)
	return net.IP(b)
}

// SetTranslation installs a NAT translation in the BPF sg_nat map
func (m *Manager) SetTranslation(t Translation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mapMgr == nil {
		return nil
	}

	key := types.FlowKey{
		SrcIP:   ipToU32(t.OrigIP),
		DstIP:   0,
		SrcPort: t.OrigPort,
		DstPort: 0,
		Proto:   t.Protocol,
	}

	entry := types.NATEntry{
		OrigIP:    ipToU32(t.OrigIP),
		OrigPort:  t.OrigPort,
		TransIP:   ipToU32(t.TransIP),
		TransPort: t.TransPort,
		Proto:     t.Protocol,
	}

	return m.mapMgr.SetNAT(key, entry)
}

// LookupTranslation queries the NAT translation for an original IP/port
func (m *Manager) LookupTranslation(origIP net.IP, origPort uint16, proto uint8) (*Translation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mapMgr == nil {
		return nil, fmt.Errorf("map manager not initialized")
	}

	key := types.FlowKey{
		SrcIP:   ipToU32(origIP),
		DstIP:   0,
		SrcPort: origPort,
		DstPort: 0,
		Proto:   proto,
	}

	entry, err := m.mapMgr.LookupNAT(key)
	if err != nil {
		return nil, err
	}

	return &Translation{
		OrigIP:    u32ToIP(entry.OrigIP),
		OrigPort:  entry.OrigPort,
		TransIP:   u32ToIP(entry.TransIP),
		TransPort: entry.TransPort,
		Protocol:  entry.Proto,
	}, nil
}

// DeleteTranslation removes a NAT rule from the dataplane
func (m *Manager) DeleteTranslation(origIP net.IP, origPort uint16, proto uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mapMgr == nil {
		return nil
	}

	key := types.FlowKey{
		SrcIP:   ipToU32(origIP),
		DstIP:   0,
		SrcPort: origPort,
		DstPort: 0,
		Proto:   proto,
	}

	return m.mapMgr.DeleteNAT(key)
}
