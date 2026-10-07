package routing

import (
	"fmt"
	"net"
	"sync"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// RouteManager bridges FIB decision logic with kernel eBPF route maps
type RouteManager struct {
	mu        sync.RWMutex
	fib       *FIBEngine
	mapMgr    *maps.MapManager
	nextHopID uint32
}

// NewRouteManager creates a RouteManager
func NewRouteManager(fib *FIBEngine, mapMgr *maps.MapManager) *RouteManager {
	return &RouteManager{
		fib:       fib,
		mapMgr:    mapMgr,
		nextHopID: 100,
	}
}

// SyncRoute installs or updates a route in both the FIB and eBPF maps
func (rm *RouteManager) SyncRoute(entry *RouteEntry) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	inserted, err := rm.fib.InsertRoute(entry)
	if err != nil {
		return err
	}
	if !inserted {
		return nil // Inferior route, no-op
	}

	if rm.mapMgr != nil {
		lpmKey, err := ToLPMV4Key(entry.Destination)
		if err != nil {
			return err
		}

		// Allocate NextHop ID if not assigned
		if entry.NextHopID == 0 {
			entry.NextHopID = rm.nextHopID
			rm.nextHopID++
		}

		// Update BPF next-hop map
		nhInfo := types.NextHopInfo{
			Flags: 1,
			VRFID: entry.VRFID,
		}
		if entry.NextHopIP != nil {
			if ip4 := entry.NextHopIP.To4(); ip4 != nil {
				nhInfo.IP = uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
			}
		}
		_ = rm.mapMgr.AddNextHop(entry.NextHopID, nhInfo)

		// Install into BPF LPM route table
		return rm.mapMgr.AddRoute(lpmKey, entry.NextHopID)
	}
	return nil
}

// WithdrawRoute removes a route from the FIB and BPF map
func (rm *RouteManager) WithdrawRoute(vrfID uint32, dst net.IPNet) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	removed := rm.fib.DeleteRoute(vrfID, dst)
	if !removed {
		return fmt.Errorf("route %s not found in VRF %d", dst.String(), vrfID)
	}

	if rm.mapMgr != nil {
		lpmKey, err := ToLPMV4Key(dst)
		if err != nil {
			return err
		}
		return rm.mapMgr.DeleteRoute(lpmKey)
	}
	return nil
}
