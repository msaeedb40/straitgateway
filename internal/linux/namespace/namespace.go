// Package namespace provides Linux network namespace management utilities.
// Used by the CNI plugin to enter pod network namespaces during ADD/DEL/CHECK.
package namespace

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/sys/unix"
)

// NetNS represents a Linux network namespace.
type NetNS struct {
	fd   int
	path string
}

// GetFromPath opens the network namespace at the given path (e.g., /var/run/netns/foo
// or /proc/<pid>/ns/net).
func GetFromPath(path string) (*NetNS, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open netns %q: %w", path, err)
	}
	return &NetNS{fd: fd, path: path}, nil
}

// GetCurrent returns the current goroutine's network namespace.
func GetCurrent() (*NetNS, error) {
	path := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid())
	return GetFromPath(path)
}

// Close closes the network namespace file descriptor.
func (n *NetNS) Close() error {
	return unix.Close(n.fd)
}

// Path returns the filesystem path of the namespace.
func (n *NetNS) Path() string { return n.path }

// Do executes fn inside the network namespace, then restores the original namespace.
// The goroutine is locked to its OS thread for the duration.
func (n *NetNS) Do(fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save current namespace
	origNS, err := GetCurrent()
	if err != nil {
		return fmt.Errorf("get current netns: %w", err)
	}
	defer origNS.Close()

	// Enter target namespace
	if err := unix.Setns(n.fd, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("setns %q: %w", n.path, err)
	}

	// Execute function
	fnErr := fn()

	// Restore original namespace
	if err := unix.Setns(origNS.fd, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("restore netns: %w", err)
	}

	return fnErr
}
