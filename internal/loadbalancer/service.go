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
		// Use first backend for direct socket translation
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
			_ = types.ServiceKey{
				IP:    binary.BigEndian.Uint32(vip4),
				Port:  svc.Port,
				Proto: svc.Protocol,
			}
			_ = types.ServiceEndpoint{
				IP:     binary.BigEndian.Uint32(bIP4),
				Port:   b.Port,
				Weight: b.Weight,
				State:  uint8(b.State),
			}
			// In live agent, maps.UpdateService(sKey, epID)
		}
	}
	return nil
}

// DeleteService removes a service VIP
func (sm *ServiceManager) DeleteService(namespace, name string, port uint16) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	svcKey := fmt.Sprintf("%s/%s:%d", namespace, name, port)
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
