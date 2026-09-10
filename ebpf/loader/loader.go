// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package loader manages the lifecycle of eBPF programs and maps.
// Uses cilium/ebpf with CO-RE (Compile Once, Run Everywhere) via BTF.
//
// Programs are loaded from ELF objects compiled by bpf2go, then attached
// to their respective hooks (NetKit/TCX/XDP/cgroup/LSM).
//
// Maps are pinned to /sys/fs/bpf/straitgateway/ for persistence across
// agent restarts.
package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
	"go.uber.org/zap"
)

const (
	// PinPath is the BPF filesystem directory for straitgateway pinned maps.
	PinPath = "/sys/fs/bpf/straitgateway"
	// ObjectDir is the directory containing compiled BPF ELF objects.
	ObjectDir = "/var/lib/straitgateway/bpf"
)

// Loader manages loading and attaching eBPF programs.
type Loader struct {
	log     *zap.Logger
	pinPath string
	objDir  string
	// collections holds loaded eBPF program/map collections.
	collections map[string]*ebpf.Collection
}

// New creates a new eBPF Loader.
func New(log *zap.Logger) *Loader {
	return &Loader{
		log:         log,
		pinPath:     PinPath,
		objDir:      ObjectDir,
		collections: make(map[string]*ebpf.Collection),
	}
}

// LoadAll loads all straitgateway eBPF programs from the object directory.
func (l *Loader) LoadAll() error {
	programs := []string{
		"netkit_pod",
		"tcx_host",
		"xdp_edge",
		"conntrack",
		"cgroup_sock",
		"lsm_net",
		"trace",
	}

	for _, prog := range programs {
		objPath := filepath.Join(l.objDir, prog+".o")
		if _, err := os.Stat(objPath); os.IsNotExist(err) {
			l.log.Warn("BPF object not found, skipping", zap.String("program", prog))
			continue
		}

		spec, err := ebpf.LoadCollectionSpec(objPath)
		if err != nil {
			return fmt.Errorf("loading BPF spec %s: %w", prog, err)
		}

		// Pin maps for persistence across restarts.
		opts := &ebpf.CollectionOptions{
			Maps: ebpf.MapOptions{
				PinPath: l.pinPath,
			},
		}

		coll, err := ebpf.NewCollectionWithOptions(spec, *opts)
		if err != nil {
			return fmt.Errorf("loading BPF collection %s: %w", prog, err)
		}

		l.collections[prog] = coll
		l.log.Info("BPF program loaded",
			zap.String("program", prog),
			zap.Int("maps", len(coll.Maps)),
			zap.Int("programs", len(coll.Programs)),
		)
	}

	return nil
}

// GetMap returns a loaded BPF map by collection and map name.
func (l *Loader) GetMap(collection, mapName string) (*ebpf.Map, error) {
	coll, ok := l.collections[collection]
	if !ok {
		return nil, fmt.Errorf("collection %q not loaded", collection)
	}
	m, ok := coll.Maps[mapName]
	if !ok {
		return nil, fmt.Errorf("map %q not found in collection %q", mapName, collection)
	}
	return m, nil
}

// Close unloads all eBPF programs and unpins maps.
func (l *Loader) Close() error {
	for name, coll := range l.collections {
		coll.Close()
		l.log.Info("BPF collection closed", zap.String("program", name))
	}
	return nil
}
