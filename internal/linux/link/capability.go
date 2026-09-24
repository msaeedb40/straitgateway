// Package link — capability.go
// Probes Linux kernel capabilities to determine whether Netkit and TCX
// are supported natively or require veth fallback.
package link

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// DatapathMode represents the selected pod interface technology.
type DatapathMode string

const (
	DatapathModeNetkit DatapathMode = "netkit"
	DatapathModeVeth   DatapathMode = "veth"
)

// Capabilities holds detected host kernel features.
type Capabilities struct {
	KernelRelease string       `json:"kernelRelease"`
	Major         int          `json:"major"`
	Minor         int          `json:"minor"`
	HasNetkit     bool         `json:"hasNetkit"`
	HasTCX        bool         `json:"hasTCX"`
	DatapathMode  DatapathMode `json:"datapathMode"`
}

// DetectCapabilities inspects /proc/version and kernel release to determine capabilities.
func DetectCapabilities() *Capabilities {
	release := readKernelRelease()
	major, minor := parseKernelVersion(release)

	// Netkit was introduced in Linux 6.7
	hasNetkit := (major > 6) || (major == 6 && minor >= 7)

	// TCX was introduced in Linux 6.6
	hasTCX := (major > 6) || (major == 6 && minor >= 6)

	mode := DatapathModeNetkit
	if !hasNetkit {
		mode = DatapathModeVeth
	}

	return &Capabilities{
		KernelRelease: release,
		Major:         major,
		Minor:         minor,
		HasNetkit:     hasNetkit,
		HasTCX:        hasTCX,
		DatapathMode:  mode,
	}
}

func readKernelRelease() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	file, err := os.Open("/proc/version")
	if err != nil {
		return "6.8.0" // safe default assumption
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			return fields[2]
		}
	}
	return "6.8.0"
}

func parseKernelVersion(release string) (int, int) {
	parts := strings.Split(release, ".")
	if len(parts) < 2 {
		return 6, 8
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 6, 8
	}

	// Minor might have trailing non-digits like 6.8.0-45-generic
	minorStr := parts[1]
	for i, c := range minorStr {
		if c < '0' || c > '9' {
			minorStr = minorStr[:i]
			break
		}
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return major, 8
	}

	return major, minor
}
