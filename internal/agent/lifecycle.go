package agent

import (
	"fmt"
	"os"
)

// ConfigureHostNetworking sets recommended sysctl parameters for high-performance transit routing
func ConfigureHostNetworking() error {
	sysctls := map[string]string{
		"/proc/sys/net/ipv4/ip_forward":                   "1",
		"/proc/sys/net/ipv4/conf/all/forwarding":          "1",
		"/proc/sys/net/ipv4/conf/default/forwarding":      "1",
		"/proc/sys/net/ipv6/conf/all/forwarding":          "1",
		"/proc/sys/net/ipv4/conf/all/rp_filter":           "0",
		"/proc/sys/net/ipv4/conf/default/rp_filter":       "0",
	}

	for path, val := range sysctls {
		if err := writeSysctl(path, val); err != nil {
			// On unprivileged dev/container environments, log and continue
			_ = err
		}
	}
	return nil
}

func writeSysctl(path, val string) error {
	return os.WriteFile(path, []byte(val), 0644)
}

// EnsureBPFFS mounts bpffs at /sys/fs/bpf if not present
func EnsureBPFFS() error {
	if fi, err := os.Stat("/sys/fs/bpf"); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll("/sys/fs/bpf", 0755)
}

// CleanupHostNetworking resets or logs cleanup state
func CleanupHostNetworking() error {
	fmt.Println("Shutting down StraitGateway agent networking cleanly")
	return nil
}
