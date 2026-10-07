package loader

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
)

// DataplaneCollection holds loaded BPF programs and their associated maps
type DataplaneCollection struct {
	mu         sync.RWMutex
	mapManager *maps.MapManager
	tcIngress  *ebpf.Program
	tcEgress   *ebpf.Program
	xdpIngress *ebpf.Program
	netkitIn   *ebpf.Program
	netkitOut  *ebpf.Program
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

// Close releases loaded resources
func (dc *DataplaneCollection) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	var errs []error
	progs := []*ebpf.Program{
		dc.tcIngress, dc.tcEgress, dc.xdpIngress,
		dc.netkitIn, dc.netkitOut,
	}
	for _, p := range progs {
		if p != nil {
			if err := p.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing programs: %v", errs)
	}
	return nil
}
