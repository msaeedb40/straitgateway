// Package link provides Netlink-based Linux network link management.
// Used by straitd to create, configure, and destroy Netkit devices
// and other Linux network interfaces.
package link

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// NetkitLink represents a Netkit device pair (parent + peer).
type NetkitLink struct {
	// Parent is the host-side Netkit interface (owned by straitd).
	Parent netlink.Link

	// Peer is the pod-side Netkit interface (moved into pod netns).
	Peer netlink.Link

	// IsNetkit indicates whether native netkit was used (false = veth fallback).
	IsNetkit bool
}

// CreateNetkit creates a Netkit parent/peer device pair for a pod.
// The parent remains in the host network namespace; the peer is moved
// into the pod network namespace by the CNI plugin.
//
// name is the base name for the parent interface (e.g., "sg-abc123").
// peerName is the name for the peer interface inside the pod (e.g., "eth0" or temp "sgpeer").
func CreateNetkit(name, peerName string, mtu int) (*NetkitLink, error) {
	if mtu <= 0 {
		mtu = 1500
	}

	// 1. Try native Netkit via 'ip link add' (kernel 6.7+)
	cmd := exec.Command("ip", "link", "add", name, "type", "netkit", "mode", "l3", "peer", "name", peerName)
	if err := cmd.Run(); err == nil {
		parent, errP := netlink.LinkByName(name)
		peer, errPeer := netlink.LinkByName(peerName)
		if errP == nil && errPeer == nil {
			_ = netlink.LinkSetMTU(parent, mtu)
			_ = netlink.LinkSetMTU(peer, mtu)
			return &NetkitLink{
				Parent:   parent,
				Peer:     peer,
				IsNetkit: true,
			}, nil
		}
	}

	// 2. Fallback to veth pair if netkit is unsupported or unprivileged
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
			MTU:  mtu,
		},
		PeerName: peerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, fmt.Errorf("failed to create link (netkit & veth fallback failed): %w", err)
	}

	parent, err := netlink.LinkByName(name)
	if err != nil {
		_ = netlink.LinkDel(veth)
		return nil, fmt.Errorf("lookup parent link %q: %w", name, err)
	}

	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		_ = netlink.LinkDel(parent)
		return nil, fmt.Errorf("lookup peer link %q: %w", peerName, err)
	}

	_ = netlink.LinkSetMTU(peer, mtu)

	return &NetkitLink{
		Parent:   parent,
		Peer:     peer,
		IsNetkit: false,
	}, nil
}

// DeleteNetkit removes a Netkit/veth device and its peer.
func DeleteNetkit(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		// Idempotent: if the link is already gone, that's fine.
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return nil
		}
		return nil
	}
	return netlink.LinkDel(link)
}

// SetLinkUp brings a network interface up.
func SetLinkUp(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("SetLinkUp: %w", err)
	}
	return netlink.LinkSetUp(link)
}

// MoveLinkToNetNS moves a link into the specified network namespace.
func MoveLinkToNetNS(link netlink.Link, netnsFD int) error {
	return netlink.LinkSetNsFd(link, netnsFD)
}

// RenameLink renames an existing network interface.
func RenameLink(oldName, newName string) error {
	link, err := netlink.LinkByName(oldName)
	if err != nil {
		return fmt.Errorf("RenameLink lookup %q: %w", oldName, err)
	}
	return netlink.LinkSetName(link, newName)
}

// AddAddr assigns an IP address to a network interface.
func AddAddr(name string, cidr string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("AddAddr: lookup %q: %w", name, err)
	}
	addr, err := netlink.ParseAddr(cidr)
	if err != nil {
		return fmt.Errorf("AddAddr: parse %q: %w", cidr, err)
	}
	return netlink.AddrAdd(link, addr)
}

// AddDefaultRoute adds a default IPv4 or IPv6 route via the given gateway on an interface.
func AddDefaultRoute(ifName string, gw net.IP) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("AddDefaultRoute: lookup %q: %w", ifName, err)
	}

	family := unix.AF_INET
	var dst *net.IPNet
	if gw.To4() == nil {
		family = unix.AF_INET6
		_, dst, _ = net.ParseCIDR("::/0")
	} else {
		_, dst, _ = net.ParseCIDR("0.0.0.0/0")
	}

	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gw,
		Dst:       dst,
		Family:    family,
	})
}
