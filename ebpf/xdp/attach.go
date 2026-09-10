// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package xdp manages XDP program attachment for NodePort acceleration
// and DDoS filtering at the NIC driver level.
package xdp

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"go.uber.org/zap"
)

// Attacher manages XDP program attachments.
type Attacher struct {
	log   *zap.Logger
	links map[int]link.Link
}

// New creates a new XDP Attacher.
func New(log *zap.Logger) *Attacher {
	return &Attacher{
		log:   log,
		links: make(map[int]link.Link),
	}
}

// Attach attaches an XDP program to the given interface.
func (a *Attacher) Attach(ifIndex int, prog *ebpf.Program) error {
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: ifIndex,
	})
	if err != nil {
		return fmt.Errorf("attaching XDP on ifindex %d: %w", ifIndex, err)
	}
	a.links[ifIndex] = l
	a.log.Info("XDP program attached", zap.Int("ifindex", ifIndex))
	return nil
}

// Detach removes the XDP program from the given interface.
func (a *Attacher) Detach(ifIndex int) error {
	if l, ok := a.links[ifIndex]; ok {
		if err := l.Close(); err != nil {
			return fmt.Errorf("detaching XDP from ifindex %d: %w", ifIndex, err)
		}
		delete(a.links, ifIndex)
	}
	return nil
}

// Close detaches all XDP programs.
func (a *Attacher) Close() error {
	for idx, l := range a.links {
		if err := l.Close(); err != nil {
			a.log.Warn("XDP detach failed", zap.Int("ifindex", idx), zap.Error(err))
		}
	}
	return nil
}
