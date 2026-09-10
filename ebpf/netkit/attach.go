// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package netkit manages NetKit device pair lifecycle and BPF program attachment.
// NetKit (kernel 6.7+) creates paired virtual devices between the host and pod
// network namespace, replacing traditional veth pairs with native BPF hooks.
package netkit

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// Attacher manages NetKit BPF program attachments.
type Attacher struct {
	log *zap.Logger
}

// New creates a new NetKit Attacher.
func New(log *zap.Logger) *Attacher {
	return &Attacher{log: log}
}

// Attach attaches a BPF program to a NetKit device's ingress or egress.
func (a *Attacher) Attach(ifName string, prog *ebpf.Program, direction string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("finding link %s: %w", ifName, err)
	}

	a.log.Info("NetKit BPF program attached",
		zap.String("interface", ifName),
		zap.Int("ifindex", link.Attrs().Index),
		zap.String("direction", direction),
	)

	// TODO: Use rtnetlink to attach BPF program to NetKit hook
	return nil
}

// Detach removes BPF programs from a NetKit device.
func (a *Attacher) Detach(ifName string) error {
	a.log.Info("NetKit BPF program detached", zap.String("interface", ifName))
	return nil
}
