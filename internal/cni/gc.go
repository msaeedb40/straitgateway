// Package cni — gc.go
// Implements garbage collection for orphaned Netkit interfaces and stale IP leases.
package cni

import (
	"fmt"
	"strings"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// GarbageCollector cleans up dangling network artifacts after container runtime failure.
type GarbageCollector struct {
	log *zap.Logger
}

// NewGarbageCollector creates a new CNI GC instance.
func NewGarbageCollector(log *zap.Logger) *GarbageCollector {
	if log == nil {
		log = zap.NewNop()
	}
	return &GarbageCollector{log: log}
}

// Collect sweeps host links, identifying orphaned sg- or nk- interfaces.
func (gc *GarbageCollector) Collect() (int, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return 0, fmt.Errorf("failed to list host links for GC: %w", err)
	}

	cleaned := 0
	for _, link := range links {
		name := link.Attrs().Name
		// Target StraitGateway interfaces: sg-* or nk-*
		if strings.HasPrefix(name, "nk-") || strings.HasPrefix(name, "sg-") {
			// Check if peer is orphaned/down
			if link.Attrs().OperState == netlink.OperDown {
				gc.log.Info("Removing orphaned interface during CNI GC", zap.String("interface", name))
				if err := netlink.LinkDel(link); err == nil {
					cleaned++
				}
			}
		}
	}

	return cleaned, nil
}
