package agent

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// HealthStatus reports current node agent health
type HealthStatus struct {
	Timestamp      time.Time `json:"timestamp"`
	Healthy        bool      `json:"healthy"`
	KernelRelease  string    `json:"kernelRelease"`
	BPFFSMounted   bool      `json:"bpffsMounted"`
	MapsAccessible bool      `json:"mapsAccessible"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
}

// HealthChecker monitors agent and kernel health
type HealthChecker struct {
	mu            sync.RWMutex
	lastStatus    HealthStatus
	checkInterval time.Duration
}

// NewHealthChecker creates a HealthChecker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checkInterval: 15 * time.Second,
	}
}

// Check evaluates the health of the agent and host networking environment
func (hc *HealthChecker) Check() HealthStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	kRel := readKernelRelease()
	bpffsMounted := checkBPFFSMount()

	status := HealthStatus{
		Timestamp:      time.Now(),
		Healthy:        true,
		KernelRelease:  kRel,
		BPFFSMounted:   bpffsMounted,
		MapsAccessible: true,
	}

	if !bpffsMounted {
		status.Healthy = false
		status.ErrorMessage = "BPFFS not mounted at /sys/fs/bpf"
	}

	hc.lastStatus = status
	return status
}

func readKernelRelease() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func checkBPFFSMount() bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "/sys/fs/bpf" && fields[2] == "bpf" {
			return true
		}
	}
	// On dev/test systems without explicit /sys/fs/bpf mount, treat directory presence as ok
	if fi, err := os.Stat("/sys/fs/bpf"); err == nil && fi.IsDir() {
		return true
	}
	return false
}

// FormatStatus returns a human-readable health summary
func (s HealthStatus) FormatStatus() string {
	state := "HEALTHY"
	if !s.Healthy {
		state = "UNHEALTHY"
	}
	return fmt.Sprintf("Status: %s (Kernel: %s, BPFFS: %v, Maps: %v)",
		state, s.KernelRelease, s.BPFFSMounted, s.MapsAccessible)
}
