// Package cgroup provides cgroup v2 eBPF program attachment via cilium/ebpf.
package cgroup

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// HookType is the cgroup BPF attach type.
type HookType int

const (
	Connect4  HookType = iota // cgroup/connect4
	Connect6                  // cgroup/connect6
	Sendmsg4                  // cgroup/sendmsg4
	Sendmsg6                  // cgroup/sendmsg6
)

// Handle wraps a live cgroup link.
type Handle struct {
	lnk link.Link
}

// Close detaches the cgroup program.
func (h *Handle) Close() error {
	if h == nil || h.lnk == nil {
		return nil
	}
	return h.lnk.Close()
}

// Attach attaches prog to cgroupPath for the given hook type.
// cgroupPath is typically "/sys/fs/cgroup" for the root cgroup v2.
func Attach(cgroupPath string, prog *ebpf.Program, hook HookType) (*Handle, error) {
	var attachType ebpf.AttachType
	switch hook {
	case Connect4:
		attachType = ebpf.AttachCGroupInet4Connect
	case Connect6:
		attachType = ebpf.AttachCGroupInet6Connect
	case Sendmsg4:
		attachType = ebpf.AttachCGroupUDP4Sendmsg
	case Sendmsg6:
		attachType = ebpf.AttachCGroupUDP6Sendmsg
	default:
		return nil, fmt.Errorf("cgroup.Attach: unknown hook type %d", hook)
	}

	lnk, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  attachType,
		Program: prog,
	})
	if err != nil {
		return nil, fmt.Errorf("cgroup.Attach(%v): %w", hook, err)
	}

	return &Handle{lnk: lnk}, nil
}

// AttachSocketLB attaches socket-level load-balancer programs to the root cgroup.
// It attaches both connect4 and sendmsg4 hooks for TCP and UDP services.
func AttachSocketLB(cgroupPath string, connect4Prog, sendmsg4Prog *ebpf.Program) ([]*Handle, error) {
	handles := make([]*Handle, 0, 2)

	h, err := Attach(cgroupPath, connect4Prog, Connect4)
	if err != nil {
		return nil, fmt.Errorf("socket LB connect4: %w", err)
	}
	handles = append(handles, h)

	h, err = Attach(cgroupPath, sendmsg4Prog, Sendmsg4)
	if err != nil {
		handles[0].Close()
		return nil, fmt.Errorf("socket LB sendmsg4: %w", err)
	}
	handles = append(handles, h)

	return handles, nil
}
