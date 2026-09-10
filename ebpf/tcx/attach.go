// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package tcx manages TCX (Traffic Control eXpress) program attachment.
// TCX replaces TC-BPF with a purpose-built attachment API (kernel 6.6+).
package tcx

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"go.uber.org/zap"
)

// Attacher manages TCX program attachments.
type Attacher struct {
	log   *zap.Logger
	links map[string]link.Link
}

// New creates a new TCX Attacher.
func New(log *zap.Logger) *Attacher {
	return &Attacher{
		log:   log,
		links: make(map[string]link.Link),
	}
}

// AttachIngress attaches a BPF program to TCX ingress on an interface.
func (a *Attacher) AttachIngress(ifIndex int, prog *ebpf.Program) error {
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifIndex,
		Program:   prog,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attaching TCX ingress on ifindex %d: %w", ifIndex, err)
	}
	key := fmt.Sprintf("tcx-ingress-%d", ifIndex)
	a.links[key] = l
	a.log.Info("TCX ingress attached", zap.Int("ifindex", ifIndex))
	return nil
}

// AttachEgress attaches a BPF program to TCX egress on an interface.
func (a *Attacher) AttachEgress(ifIndex int, prog *ebpf.Program) error {
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifIndex,
		Program:   prog,
		Attach:    ebpf.AttachTCXEgress,
	})
	if err != nil {
		return fmt.Errorf("attaching TCX egress on ifindex %d: %w", ifIndex, err)
	}
	key := fmt.Sprintf("tcx-egress-%d", ifIndex)
	a.links[key] = l
	a.log.Info("TCX egress attached", zap.Int("ifindex", ifIndex))
	return nil
}

// Close detaches all TCX programs.
func (a *Attacher) Close() error {
	for key, l := range a.links {
		if err := l.Close(); err != nil {
			a.log.Warn("TCX detach failed", zap.String("key", key), zap.Error(err))
		}
	}
	return nil
}
