package loader

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/vishvananda/netlink"
)

// DataplaneCollection holds loaded BPF programs, links, and their associated maps
type DataplaneCollection struct {
	mu         sync.RWMutex
	mapManager *maps.MapManager
	tcIngress  *ebpf.Program
	tcEgress   *ebpf.Program
	xdpIngress *ebpf.Program
	netkitIn   *ebpf.Program
	netkitOut  *ebpf.Program
	cgroupCon4 *ebpf.Program
	cgroupCon6 *ebpf.Program
	cgroupSnd4 *ebpf.Program
	cgroupSnd6 *ebpf.Program
	activeLinks []link.Link
}

// LoadFromBytes loads an eBPF ELF program binary into an ebpf.Program
func LoadFromBytes(bpfELF []byte, secName string) (*ebpf.Program, error) {
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(bpfELF))
	if err != nil {
		return nil, fmt.Errorf("load collection spec: %w", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, fmt.Errorf("new collection: %w", err)
	}
	defer coll.Close()

	prog, exists := coll.Programs[secName]
	if !exists {
		return nil, fmt.Errorf("program section %s not found in ELF", secName)
	}

	// Clone the program so it survives collection close
	clonedProg, err := prog.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone program: %w", err)
	}
	return clonedProg, nil
}

// LoadFromFile loads an eBPF ELF object from disk
func LoadFromFile(filePath string, progName string) (*ebpf.Program, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read BPF ELF %s: %w", filePath, err)
	}
	return LoadFromBytes(data, progName)
}

// Loader manages loading and attaching dataplane programs
type Loader struct {
	mapMgr *maps.MapManager
}

// NewLoader creates a loader referencing the given MapManager
func NewLoader(mapMgr *maps.MapManager) *Loader {
	return &Loader{
		mapMgr: mapMgr,
	}
}

// AttachTCXIngress attaches a TCX ingress program to the specified interface index (Linux 6.6+)
func (l *Loader) AttachTCXIngress(ifIndex int, prog *ebpf.Program) (link.Link, error) {
	tcxLink, err := link.AttachTCX(link.TCXOptions{
		Interface: ifIndex,
		Program:   prog,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return nil, fmt.Errorf("attach tcx ingress on ifindex %d: %w", ifIndex, err)
	}
	return tcxLink, nil
}

// AttachTCXEgress attaches a TCX egress program to the specified interface index (Linux 6.6+)
func (l *Loader) AttachTCXEgress(ifIndex int, prog *ebpf.Program) (link.Link, error) {
	tcxLink, err := link.AttachTCX(link.TCXOptions{
		Interface: ifIndex,
		Program:   prog,
		Attach:    ebpf.AttachTCXEgress,
	})
	if err != nil {
		return nil, fmt.Errorf("attach tcx egress on ifindex %d: %w", ifIndex, err)
	}
	return tcxLink, nil
}

// AttachXDP attaches an XDP program to an interface (driver or generic fallback)
func (l *Loader) AttachXDP(ifIndex int, prog *ebpf.Program, flags link.XDPAttachFlags) (link.Link, error) {
	xdpLink, err := link.AttachXDP(link.XDPOptions{
		Interface: ifIndex,
		Program:   prog,
		Flags:     flags,
	})
	if err != nil {
		return nil, fmt.Errorf("attach xdp on ifindex %d: %w", ifIndex, err)
	}
	return xdpLink, nil
}

// AttachCgroupSocketLB attaches socket-level load balancing programs to cgroup v2
func (l *Loader) AttachCgroupSocketLB(
	cgroupPath string,
	connect4 *ebpf.Program,
	connect6 *ebpf.Program,
	sendmsg4 *ebpf.Program,
	sendmsg6 *ebpf.Program,
) ([]link.Link, error) {
	var links []link.Link

	cleanupOnError := func() {
		for _, lk := range links {
			_ = lk.Close()
		}
	}

	if connect4 != nil {
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Program: connect4,
			Attach:  ebpf.AttachCGroupInet4Connect,
		})
		if err != nil {
			cleanupOnError()
			return nil, fmt.Errorf("attach cgroup connect4: %w", err)
		}
		links = append(links, lk)
	}

	if connect6 != nil {
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Program: connect6,
			Attach:  ebpf.AttachCGroupInet6Connect,
		})
		if err != nil {
			cleanupOnError()
			return nil, fmt.Errorf("attach cgroup connect6: %w", err)
		}
		links = append(links, lk)
	}

	if sendmsg4 != nil {
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Program: sendmsg4,
			Attach:  ebpf.AttachCGroupUDP4Sendmsg,
		})
		if err != nil {
			cleanupOnError()
			return nil, fmt.Errorf("attach cgroup sendmsg4: %w", err)
		}
		links = append(links, lk)
	}

	if sendmsg6 != nil {
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Program: sendmsg6,
			Attach:  ebpf.AttachCGroupUDP6Sendmsg,
		})
		if err != nil {
			cleanupOnError()
			return nil, fmt.Errorf("attach cgroup sendmsg6: %w", err)
		}
		links = append(links, lk)
	}

	return links, nil
}

// AttachNetkit attaches ingress and egress TCX hooks to a netkit interface pair
func (l *Loader) AttachNetkit(primaryIndex, peerIndex int, ingressProg, egressProg *ebpf.Program) ([]link.Link, error) {
	var links []link.Link

	if ingressProg != nil && primaryIndex > 0 {
		inLink, err := l.AttachTCXIngress(primaryIndex, ingressProg)
		if err != nil {
			return nil, fmt.Errorf("netkit primary ingress hook: %w", err)
		}
		links = append(links, inLink)
	}

	if egressProg != nil && peerIndex > 0 {
		outLink, err := l.AttachTCXEgress(peerIndex, egressProg)
		if err != nil {
			for _, lk := range links {
				_ = lk.Close()
			}
			return nil, fmt.Errorf("netkit peer egress hook: %w", err)
		}
		links = append(links, outLink)
	}

	return links, nil
}

// GetInterfaceIndexByName resolves a network interface name to its kernel ifindex via netlink
func GetInterfaceIndexByName(ifName string) (int, error) {
	nlLink, err := netlink.LinkByName(ifName)
	if err != nil {
		return 0, fmt.Errorf("find interface %s: %w", ifName, err)
	}
	return nlLink.Attrs().Index, nil
}

// AddLink registers an active link to be closed when collection terminates
func (dc *DataplaneCollection) AddLink(l link.Link) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.activeLinks = append(dc.activeLinks, l)
}

// Close releases all loaded programs, links, and map handles
func (dc *DataplaneCollection) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	var errs []error

	// 1. Detach links first
	for _, lk := range dc.activeLinks {
		if lk != nil {
			if err := lk.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	dc.activeLinks = nil

	// 2. Close loaded programs
	progs := []*ebpf.Program{
		dc.tcIngress, dc.tcEgress, dc.xdpIngress,
		dc.netkitIn, dc.netkitOut,
		dc.cgroupCon4, dc.cgroupCon6,
		dc.cgroupSnd4, dc.cgroupSnd6,
	}
	for _, p := range progs {
		if p != nil {
			if err := p.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing dataplane collection: %v", errs)
	}
	return nil
}
