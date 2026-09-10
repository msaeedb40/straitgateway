// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package bpf provides versioned BPF map pin/unpin helpers.
package bpf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
)

const (
	// DefaultPinRoot is the bpffs directory for straitgateway pinned maps.
	DefaultPinRoot = "/sys/fs/bpf/straitgateway"
)

// PinMap pins an eBPF map to the bpffs with a versioned path.
// Path format: /sys/fs/bpf/straitgateway/<mapName>
func PinMap(m *ebpf.Map, name string) error {
	path := filepath.Join(DefaultPinRoot, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating pin directory: %w", err)
	}
	// Remove stale pin if it exists.
	_ = os.Remove(path)
	if err := m.Pin(path); err != nil {
		return fmt.Errorf("pinning map %s to %s: %w", name, path, err)
	}
	return nil
}

// UnpinMap removes a pinned map from bpffs.
func UnpinMap(name string) error {
	path := filepath.Join(DefaultPinRoot, name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unpinning map %s: %w", name, err)
	}
	return nil
}

// LoadPinnedMap loads an existing pinned map from bpffs.
func LoadPinnedMap(name string) (*ebpf.Map, error) {
	path := filepath.Join(DefaultPinRoot, name)
	m, err := ebpf.LoadPinnedMap(path, nil)
	if err != nil {
		return nil, fmt.Errorf("loading pinned map %s: %w", path, err)
	}
	return m, nil
}

// CleanupAll removes all straitgateway pinned maps.
func CleanupAll() error {
	return os.RemoveAll(DefaultPinRoot)
}
