// Package runtime provides container runtime environment detection and netns helpers for CNI.
package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RuntimeType identifies the container runtime.
type RuntimeType string

const (
	RuntimeContainerd RuntimeType = "containerd"
	RuntimeCRIO       RuntimeType = "crio"
	RuntimeDocker     RuntimeType = "docker"
	RuntimeUnknown    RuntimeType = "unknown"
)

// DetectRuntime probes active container runtime sockets on the host.
func DetectRuntime() RuntimeType {
	sockets := []struct {
		path string
		rt   RuntimeType
	}{
		{"/run/containerd/containerd.sock", RuntimeContainerd},
		{"/var/run/dockershim.sock", RuntimeDocker},
		{"/run/crio/crio.sock", RuntimeCRIO},
	}

	for _, s := range sockets {
		if _, err := os.Stat(s.path); err == nil {
			return s.rt
		}
	}
	return RuntimeContainerd
}

// ExtractPIDFromNetNS extracts the process ID from a netns path like /proc/<pid>/ns/net.
func ExtractPIDFromNetNS(netnsPath string) (int, error) {
	parts := strings.Split(strings.Trim(netnsPath, "/"), "/")
	if len(parts) >= 2 && parts[0] == "proc" && parts[2] == "ns" {
		pid, err := strconv.Atoi(parts[1])
		if err == nil && pid > 0 {
			return pid, nil
		}
	}
	return 0, fmt.Errorf("path %q is not in /proc/<pid>/ns/net format", netnsPath)
}
