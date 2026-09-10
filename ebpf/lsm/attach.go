// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package lsm manages Linux Security Module (LSM) BPF hook attachments.
// LSM hooks provide kernel-level policy enforcement that cannot be bypassed.
package lsm

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"go.uber.org/zap"
)

// Attacher manages LSM BPF hook attachments.
type Attacher struct {
	log   *zap.Logger
	links []link.Link
}

// New creates a new LSM Attacher.
func New(log *zap.Logger) *Attacher {
	return &Attacher{log: log}
}

// Attach attaches a BPF program to an LSM hook.
func (a *Attacher) Attach(prog *ebpf.Program) error {
	l, err := link.AttachLSM(link.LSMOptions{
		Program: prog,
	})
	if err != nil {
		return fmt.Errorf("attaching LSM BPF: %w", err)
	}
	a.links = append(a.links, l)
	a.log.Info("LSM BPF hook attached")
	return nil
}

// Close detaches all LSM hooks.
func (a *Attacher) Close() error {
	for _, l := range a.links {
		l.Close()
	}
	return nil
}
