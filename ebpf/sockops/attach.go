// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package sockops manages sockops and cgroup BPF program attachment.
// Used for socket-level service redirect and identity tagging.
package sockops

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"go.uber.org/zap"
)

// Attacher manages cgroup BPF program attachments.
type Attacher struct {
	log   *zap.Logger
	links []link.Link
}

// New creates a new sockops/cgroup Attacher.
func New(log *zap.Logger) *Attacher {
	return &Attacher{log: log}
}

// AttachCgroup attaches a BPF program to a cgroup for socket operations.
func (a *Attacher) AttachCgroup(cgroupPath string, prog *ebpf.Program, attachType ebpf.AttachType) error {
	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Program: prog,
		Attach:  attachType,
	})
	if err != nil {
		return fmt.Errorf("attaching cgroup BPF to %s: %w", cgroupPath, err)
	}
	a.links = append(a.links, l)
	a.log.Info("cgroup BPF attached", zap.String("cgroup", cgroupPath))
	return nil
}

// Close detaches all cgroup BPF programs.
func (a *Attacher) Close() error {
	for _, l := range a.links {
		l.Close()
	}
	return nil
}
