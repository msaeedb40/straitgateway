// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package process provides Linux kernel and process utilities for straitgatewayd.
package process

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

// MinKernelVersion is the minimum supported Linux kernel version.
// NetKit requires 6.7+, BTF CO-RE requires 5.8+.
const MinKernelVersion = "6.7.0"

// KernelVersion holds a parsed kernel version.
type KernelVersion struct {
	Major, Minor, Patch int
}

// ParseKernelVersion parses a kernel version string like "6.7.3".
func ParseKernelVersion(s string) (KernelVersion, error) {
	// Strip "-" suffix (e.g. "6.7.0-1-generic")
	if idx := strings.IndexByte(s, '-'); idx > 0 {
		s = s[:idx]
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return KernelVersion{}, fmt.Errorf("cannot parse kernel version %q", s)
	}
	if len(parts) == 2 {
		parts = append(parts, "0")
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return KernelVersion{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return KernelVersion{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return KernelVersion{}, err
	}
	return KernelVersion{Major: major, Minor: minor, Patch: patch}, nil
}

// AtLeast returns true if kv >= other.
func (kv KernelVersion) AtLeast(other KernelVersion) bool {
	if kv.Major != other.Major {
		return kv.Major > other.Major
	}
	if kv.Minor != other.Minor {
		return kv.Minor > other.Minor
	}
	return kv.Patch >= other.Patch
}

// GetRunningKernelVersion returns the currently running kernel version.
func GetRunningKernelVersion() (KernelVersion, error) {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return KernelVersion{}, fmt.Errorf("uname: %w", err)
	}
	// uname.Release is int8 array
	b := make([]byte, 0, len(uname.Release))
	for _, c := range uname.Release {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return ParseKernelVersion(string(b))
}

// CheckKernelVersion verifies the running kernel meets the minimum requirement.
func CheckKernelVersion(log *zap.Logger) error {
	kv, err := GetRunningKernelVersion()
	if err != nil {
		return fmt.Errorf("detecting kernel version: %w", err)
	}

	min, _ := ParseKernelVersion(MinKernelVersion)

	log.Info("kernel version check",
		zap.String("running", fmt.Sprintf("%d.%d.%d", kv.Major, kv.Minor, kv.Patch)),
		zap.String("minimum", MinKernelVersion),
	)

	if !kv.AtLeast(min) {
		return fmt.Errorf("kernel %d.%d.%d is below minimum required %s (NetKit requires 6.7+)",
			kv.Major, kv.Minor, kv.Patch, MinKernelVersion)
	}
	return nil
}

// EnsureBPFFS ensures the BPF filesystem is mounted and the straitgateway
// pin directory exists.
func EnsureBPFFS(bpffsPath string) error {
	// Check that bpffs is mounted.
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/sys/fs/bpf", &stat); err != nil {
		return fmt.Errorf("bpffs not mounted at /sys/fs/bpf: %w", err)
	}

	// 0xcafe4a11 is the BPF filesystem magic number.
	const bpfFSMagic = 0xcafe4a11
	if stat.Type != bpfFSMagic {
		return fmt.Errorf("/sys/fs/bpf is not a BPF filesystem (type=0x%x)", stat.Type)
	}

	// Create pin directory.
	if err := os.MkdirAll(bpffsPath, 0700); err != nil {
		return fmt.Errorf("creating BPF pin dir %s: %w", bpffsPath, err)
	}

	return nil
}
