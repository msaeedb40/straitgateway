// Package route provides Linux FIB route management for StraitGateway.
package route

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

// Add adds an IPv4 route.
func Add(dst *net.IPNet, gw net.IP, ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("route.Add: interface %q: %w", ifName, err)
	}
	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Gw:        gw,
	})
}

// Del removes an IPv4 route.
func Del(dst *net.IPNet, ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("route.Del: interface %q: %w", ifName, err)
	}
	return netlink.RouteDel(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
	})
}
