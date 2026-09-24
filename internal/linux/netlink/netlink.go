// Package netlink provides netlink-based abstractions for StraitGateway.
// Wraps github.com/vishvananda/netlink with StraitGateway-specific helpers.
package netlink

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// LinkExists returns true if the named interface exists in the current netns.
func LinkExists(name string) bool {
	_, err := netlink.LinkByName(name)
	return err == nil
}

// EnsureLinkUp brings the named interface up if it is not already.
func EnsureLinkUp(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("EnsureLinkUp: %w", err)
	}
	if link.Attrs().Flags&1 != 0 { // IFF_UP
		return nil
	}
	return netlink.LinkSetUp(link)
}

// SetMTU sets the MTU on a named interface.
func SetMTU(name string, mtu int) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("SetMTU: %w", err)
	}
	return netlink.LinkSetMTU(link, mtu)
}
