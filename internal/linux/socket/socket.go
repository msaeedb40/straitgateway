// Package socket provides helpers for socket-level eBPF operations.
package socket

import "fmt"

// AttachSockops attaches an eBPF sockops program to a cgroup.
func AttachSockops(cgroupPath string) error {
	// TODO: Use cilium/ebpf Link API for sock_ops attachment.
	return fmt.Errorf("socket.AttachSockops: not yet implemented")
}

// DetachSockops removes a sockops attachment.
func DetachSockops(cgroupPath string) error {
	return fmt.Errorf("socket.DetachSockops: not yet implemented")
}
