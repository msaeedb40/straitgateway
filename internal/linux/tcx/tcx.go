// Package tcx provides TCX attachment for Netkit pod interfaces via cilium/ebpf.
package tcx

import (
	"errors"
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
)

// Direction is the TCX attachment direction.
type Direction int

const (
	// Ingress — packets arriving at the Netkit parent (pod egress toward node).
	Ingress Direction = iota
	// Egress — packets leaving the Netkit parent toward the pod (pod ingress).
	Egress
)

// Handle wraps a live TCX link so the caller can close/detach it.
type Handle struct {
	lnk link.Link
}

// Close detaches the TCX program.
func (h *Handle) Close() error {
	if h == nil || h.lnk == nil {
		return nil
	}
	return h.lnk.Close()
}

// Attach attaches prog to ifName via TCX in the given direction.
// prog must be a loaded *ebpf.Program of type SchedCLS.
// Returns a Handle that must be closed to detach.
func Attach(ifName string, prog *ebpf.Program, dir Direction) (*Handle, error) {
	iface, err := netlink.LinkByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("tcx.Attach: interface %q: %w", ifName, err)
	}

	var lnk link.Link
	opts := link.TCXOptions{
		Interface: iface.Attrs().Index,
		Program:   prog,
		Anchor:    link.Tail(),
	}

	switch dir {
	case Ingress:
		lnk, err = link.AttachTCX(link.TCXOptions{
			Interface: opts.Interface,
			Program:   opts.Program,
			Attach:    ebpf.AttachTCXIngress,
			Anchor:    opts.Anchor,
		})
	case Egress:
		lnk, err = link.AttachTCX(link.TCXOptions{
			Interface: opts.Interface,
			Program:   opts.Program,
			Attach:    ebpf.AttachTCXEgress,
			Anchor:    opts.Anchor,
		})
	default:
		return nil, fmt.Errorf("tcx.Attach: unknown direction %d", dir)
	}

	if err != nil {
		return nil, fmt.Errorf("tcx.Attach: %w", err)
	}

	return &Handle{lnk: lnk}, nil
}

// AttachIngress is a convenience wrapper for Attach with Ingress direction.
func AttachIngress(ifName string, prog *ebpf.Program) (*Handle, error) {
	return Attach(ifName, prog, Ingress)
}

// AttachEgress is a convenience wrapper for Attach with Egress direction.
func AttachEgress(ifName string, prog *ebpf.Program) (*Handle, error) {
	return Attach(ifName, prog, Egress)
}

// IsNotSupported returns true if the error indicates the kernel does not
// support TCX (requires kernel ≥ 6.6).
func IsNotSupported(err error) bool {
	return errors.Is(err, ebpf.ErrNotSupported)
}
