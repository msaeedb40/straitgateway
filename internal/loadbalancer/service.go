package loadbalancer

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// ServiceEntry represents an active ClusterIP VIP mapping
type ServiceEntry struct {
	Namespace string
	Name      string
	VIP       net.IP
	Port      uint16
	Protocol  uint8 // 6=TCP, 17=UDP
	Backends  []Backend
	Maglev    *Maglev
}

// ServiceManager manages service VIPs and synchronizes with BPF maps
type ServiceManager struct {
	mu          sync.RWMutex
	mapMgr      *maps.MapManager
	services    map[string]*ServiceEntry // key: namespace/name:port
	nextEpID    uint32
	endpointMap map[string]uint32 // key: ip:port -> epID
}

// NewServiceManager creates a ServiceManager
func NewServiceManager(mapMgr *maps.MapManager) *ServiceManager {
	return &ServiceManager{
		mapMgr:      mapMgr,
		services:    make(map[string]*ServiceEntry),
		nextEpID:    1,
		endpointMap: make(map[string]uint32),
	}
}

// UpsertService registers or updates a service VIP and its backends
func (sm *ServiceManager) UpsertService(svc *ServiceEntry) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	svcKey := fmt.Sprintf("%s/%s:%d", svc.Namespace, svc.Name, svc.Port)

	// Precompute Maglev if backends exist
	if len(svc.Backends) > 0 {
		backendKeys := make([]string, len(svc.Backends))
		for i, b := range svc.Backends {
			backendKeys[i] = fmt.Sprintf("%s:%d", b.IP.String(), b.Port)
		}
		mag, err := NewMaglev(backendKeys, 257) // 257 prime for test/default table
		if err == nil {
			svc.Maglev = mag
		}
	}

	sm.services[svcKey] = svc

	// If MapManager is available, sync to BPF maps
	if sm.mapMgr != nil && len(svc.Backends) > 0 {
		// Use first backend as default primary endpoint for socket LB
		b := svc.Backends[0]
		bKey := fmt.Sprintf("%s:%d", b.IP.String(), b.Port)
		epID, exists := sm.endpointMap[bKey]
		if !exists {
			epID = sm.nextEpID
			sm.nextEpID++
			sm.endpointMap[bKey] = epID
		}

		vip4 := svc.VIP.To4()
		bIP4 := b.IP.To4()
		if vip4 != nil && bIP4 != nil {
			sKey := types.ServiceKey{
				IP:    binary.BigEndian.Uint32(vip4),
				Port:  svc.Port,
				Proto: svc.Protocol,
			}
			sEp := types.ServiceEndpoint{
				IP:     binary.BigEndian.Uint32(bIP4),
				Port:   b.Port,
				Weight: b.Weight,
				State:  uint8(b.State),
			}
			_ = sm.mapMgr.AddEndpoint(epID, sEp)
			_ = sm.mapMgr.AddService(sKey, epID)
		}
	}
	return nil
}

// SelectBackend selects a backend for a flow hash using Maglev consistent hashing
func (sm *ServiceManager) SelectBackend(namespace, name string, port uint16, flowHash uint64) (*Backend, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	svcKey := fmt.Sprintf("%s/%s:%d", namespace, name, port)
	svc, exists := sm.services[svcKey]
	if !exists {
		return nil, fmt.Errorf("service %s not found", svcKey)
	}

	if len(svc.Backends) == 0 {
		return nil, fmt.Errorf("no backends available for service %s", svcKey)
	}

	if svc.Maglev != nil {
		idx := svc.Maglev.Lookup(flowHash)
		if idx >= 0 && idx < len(svc.Backends) {
			return &svc.Backends[idx], nil
		}
	}

	// Fallback to modulo selection if Maglev table is pending
	idx := int(flowHash % uint64(len(svc.Backends)))
	return &svc.Backends[idx], nil
}

// DeleteService removes a service VIP and purges BPF map state
func (sm *ServiceManager) DeleteService(namespace, name string, port uint16) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	svcKey := fmt.Sprintf("%s/%s:%d", namespace, name, port)
	svc, exists := sm.services[svcKey]
	if !exists {
		return nil
	}

	if sm.mapMgr != nil {
		vip4 := svc.VIP.To4()
		if vip4 != nil {
			sKey := types.ServiceKey{
				IP:    binary.BigEndian.Uint32(vip4),
				Port:  svc.Port,
				Proto: svc.Protocol,
			}
			_ = sm.mapMgr.DeleteService(sKey)
		}
	}

	delete(sm.services, svcKey)
	return nil
}

// GetService retrieves a service entry
func (sm *ServiceManager) GetService(namespace, name string, port uint16) (*ServiceEntry, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	svcKey := fmt.Sprintf("%s/%s:%d", namespace, name, port)
	svc, exists := sm.services[svcKey]
	return svc, exists
}
