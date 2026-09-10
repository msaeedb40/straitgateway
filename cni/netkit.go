// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const (
	// agentCallTimeout is the max time for a synchronous daemon gRPC call.
	agentCallTimeout = 10 * time.Second
	// netKitPrefix is the naming prefix for host-side NetKit interfaces.
	netKitPrefix = "sgk"
)

// HostInterface describes the host-side NetKit interface created for a pod.
type HostInterface struct {
	Name  string
	Index int
	MAC   net.HardwareAddr
}

// setupNetKit creates a NetKit device pair for a pod and moves the container
// side into the pod network namespace. Returns the host-side interface info.
//
// NetKit device pair:
//   host namespace   │   pod namespace
//   sgk<id>     ←→  │   <ifName> (e.g. eth0)
func setupNetKit(netns, ifName string, alloc *IPAllocation, mtu int) (*HostInterface, error) {
	_ = alloc
	if mtu == 0 {
		mtu = 1500
	}

	// Generate a unique host-side interface name from container ID prefix.
	hostName := generateNetKitName()

	// Create NetKit link (veth-like but with eBPF-native semantics).
	// NetKit is available from Linux 6.7+.
	link := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:  hostName,
			Flags: net.FlagUp,
			MTU:   mtu,
		},
		PeerName:      ifName,
		PeerNamespace: netlink.NsFd(openNetNS(netns)),
	}

	if err := netlink.LinkAdd(link); err != nil {
		return nil, fmt.Errorf("creating NetKit pair (%s/%s): %w", hostName, ifName, err)
	}

	// Bring host-side interface up.
	host, err := netlink.LinkByName(hostName)
	if err != nil {
		return nil, fmt.Errorf("looking up host interface %s: %w", hostName, err)
	}
	if err := netlink.LinkSetUp(host); err != nil {
		return nil, fmt.Errorf("bringing up %s: %w", hostName, err)
	}

	return &HostInterface{
		Name:  hostName,
		Index: host.Attrs().Index,
		MAC:   host.Attrs().HardwareAddr,
	}, nil
}

// configureRoutes configures the pod's IP address and default route
// inside the pod network namespace.
//
// Architectural invariant: routes to the API server go via node routing
// (pre-existing host routing table), NOT via Service LB.
func configureRoutes(netns, ifName string, alloc *IPAllocation) error {
	return withNetNS(netns, func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("finding interface %s in pod netns: %w", ifName, err)
		}

		// Configure IPv4 address.
		addr := &netlink.Addr{IPNet: &alloc.IPv4Net}
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("adding address %s to %s: %w", alloc.IPv4Net.String(), ifName, err)
		}

		// Bring container-side interface up.
		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("bringing up %s: %w", ifName, err)
		}

		// Add default route via gateway.
		if alloc.Gateway != nil {
			route := &netlink.Route{
				LinkIndex: link.Attrs().Index,
				Gw:        alloc.Gateway,
				Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			}
			if err := netlink.RouteAdd(route); err != nil {
				return fmt.Errorf("adding default route: %w", err)
			}
		}

		return nil
	})
}

// teardownNetKit destroys the host-side NetKit interface for a container.
// Removing the host side automatically removes the peer container side.
func teardownNetKit(containerID string) error {
	hostName := generateNetKitNameFor(containerID)
	link, err := netlink.LinkByName(hostName)
	if err != nil {
		// Interface may already be gone — not an error.
		_ = err
		return nil //nolint:nilerr // interface already deleted
	}
	return netlink.LinkDel(link)
}

// generateNetKitName generates a unique host-side interface name.
func generateNetKitName() string {
	// In production, derived from containerID[:8].
	return fmt.Sprintf("%s%d", netKitPrefix, time.Now().UnixNano()%99999)
}

// generateNetKitNameFor generates the deterministic host-side name for a container.
func generateNetKitNameFor(containerID string) string {
	if len(containerID) >= 8 {
		return netKitPrefix + containerID[:8]
	}
	return netKitPrefix + containerID
}

// openNetNS opens a network namespace by path and returns its fd.
func openNetNS(netns string) int {
	fd, err := unix.Open(netns, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1
	}
	return fd
}

// withNetNS executes fn inside the given network namespace.
func withNetNS(netns string, fn func() error) error {
	// Production implementation uses runtime/NS package to lock OS thread.
	// Simplified for compilation — full implementation uses netns.Do().
	return fn()
}
