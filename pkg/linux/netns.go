// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package linux

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/unix"
)

// WithNetNS executes fn inside the specified network namespace.
// The calling goroutine is locked to its OS thread for the duration.
func WithNetNS(nsPath string, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current namespace.
	origNS, err := unix.Open("/proc/self/ns/net", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("opening current netns: %w", err)
	}
	defer func() { _ = unix.Close(origNS) }()

	// Open target namespace.
	targetNS, err := unix.Open(nsPath, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("opening target netns %s: %w", nsPath, err)
	}
	defer func() { _ = unix.Close(targetNS) }()

	// Enter target namespace.
	if err := unix.Setns(targetNS, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("setns to %s: %w", nsPath, err)
	}

	// Execute function, then restore original namespace.
	fnErr := fn()

	if err := unix.Setns(origNS, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("restoring original netns: %w", err)
	}

	return fnErr
}
