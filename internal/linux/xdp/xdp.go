// Package xdp provides XDP program attachment via cilium/ebpf Link API.
package xdp

import (
	"errors"
	"fmt"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// Mode is the XDP attachment mode.
type Mode int

const (
	ModeNative  Mode = iota // Driver/native XDP — fastest, requires driver support
	ModeGeneric             // Generic/SKB XDP — fallback, works on any interface
)

// Handle wraps a live XDP link.
type Handle struct {
	lnk link.Link
}

// Close detaches the XDP program.
func (h *Handle) Close() error {
	if h == nil || h.lnk == nil {
		return nil
	}
	return h.lnk.Close()
}

// Attach attaches prog to the named interface via XDP.
// prog must be a loaded *ebpf.Program of type XDP.
func Attach(ifName string, prog *ebpf.Program, mode Mode) (*Handle, error) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("xdp.Attach: interface %q: %w", ifName, err)
	}

	opts := link.XDPOptions{
		Program:   prog,
		Interface: iface.Index,
	}
	switch mode {
	case ModeGeneric:
		opts.Flags = link.XDPGenericMode
	case ModeNative:
		opts.Flags = link.XDPDriverMode
	}

	lnk, err := link.AttachXDP(opts)
	if err != nil {
		if IsNotSupported(err) && mode == ModeNative {
			// Auto-fallback to generic mode
			opts.Flags = link.XDPGenericMode
			lnk, err = link.AttachXDP(opts)
			if err != nil {
				return nil, fmt.Errorf("xdp.Attach (generic fallback): %w", err)
			}
		} else {
			return nil, fmt.Errorf("xdp.Attach: %w", err)
		}
	}

	return &Handle{lnk: lnk}, nil
}

// AttachNative attaches with native (driver) mode, falling back to generic on unsupported hardware.
func AttachNative(ifName string, prog *ebpf.Program) (*Handle, error) {
	return Attach(ifName, prog, ModeNative)
}

// AttachGeneric attaches with generic (SKB) mode.
func AttachGeneric(ifName string, prog *ebpf.Program) (*Handle, error) {
	return Attach(ifName, prog, ModeGeneric)
}

// IsNotSupported returns true if the error indicates missing XDP support.
func IsNotSupported(err error) bool {
	return errors.Is(err, ebpf.ErrNotSupported)
}
