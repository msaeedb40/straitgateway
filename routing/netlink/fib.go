// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package netlinkrouter manages the kernel FIB (Forwarding Information Base)
// via netlink. Handles route addition, deletion, ECMP multi-path, and
// policy routing tables.
//
// Architectural invariant: FIB writes only occur from the dataplane compiler.
package netlinkrouter

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// Router manages kernel FIB routes via netlink.
type Router struct {
	log *zap.Logger
}

// New creates a new FIB Router.
func New(log *zap.Logger) *Router {
	return &Router{log: log}
}

// AddRoute adds a single route to the kernel FIB.
func (r *Router) AddRoute(dst *net.IPNet, gw net.IP, ifIndex int, table int) error {
	route := &netlink.Route{
		Dst:       dst,
		Gw:        gw,
		LinkIndex: ifIndex,
		Table:     table,
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("adding route %s via %s: %w", dst, gw, err)
	}
	r.log.Debug("route added",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()),
		zap.Int("table", table),
	)
	return nil
}

// DeleteRoute removes a route from the kernel FIB.
func (r *Router) DeleteRoute(dst *net.IPNet, table int) error {
	route := &netlink.Route{
		Dst:   dst,
		Table: table,
	}
	if err := netlink.RouteDel(route); err != nil {
		return fmt.Errorf("deleting route %s: %w", dst, err)
	}
	r.log.Debug("route deleted", zap.String("dst", dst.String()))
	return nil
}

// AddECMPRoute adds a multi-path ECMP route with multiple next-hops.
func (r *Router) AddECMPRoute(dst *net.IPNet, nextHops []net.IP, table int) error {
	route := &netlink.Route{
		Dst:   dst,
		Table: table,
	}

	for _, nh := range nextHops {
		route.MultiPath = append(route.MultiPath, &netlink.NexthopInfo{
			Gw: nh,
		})
	}

	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("adding ECMP route %s: %w", dst, err)
	}
	r.log.Debug("ECMP route added",
		zap.String("dst", dst.String()),
		zap.Int("nextHops", len(nextHops)),
	)
	return nil
}
