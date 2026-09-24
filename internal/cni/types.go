package cni

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CNIArgs parses common Kubernetes CNI arguments passed via CNI_ARGS.
type CNIArgs struct {
	PodName      string
	PodNamespace string
	ContainerID  string
	PodCIDR      string
}

// CNIEnv stores the standard environment variables injected by the container runtime.
type CNIEnv struct {
	Command     string
	ContainerID string
	NetNS       string
	IfName      string
	Args        string
	Path        string
	ParsedArgs  CNIArgs
}

// ParseCNIEnv reads CNI environment variables from the process environment.
func ParseCNIEnv() (*CNIEnv, error) {
	cmd := os.Getenv("CNI_COMMAND")
	if cmd == "" {
		return nil, fmt.Errorf("CNI_COMMAND environment variable is missing")
	}

	containerID := os.Getenv("CNI_CONTAINERID")
	netns := os.Getenv("CNI_NETNS")
	ifName := os.Getenv("CNI_IFNAME")
	if ifName == "" {
		ifName = "eth0"
	}

	env := &CNIEnv{
		Command:     cmd,
		ContainerID: containerID,
		NetNS:       netns,
		IfName:      ifName,
		Args:        os.Getenv("CNI_ARGS"),
		Path:        os.Getenv("CNI_PATH"),
	}

	// Parse CNI_ARGS (key=value;key=value)
	if env.Args != "" {
		pairs := strings.Split(env.Args, ";")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				switch kv[0] {
				case "K8S_POD_NAME":
					env.ParsedArgs.PodName = kv[1]
				case "K8S_POD_NAMESPACE":
					env.ParsedArgs.PodNamespace = kv[1]
				case "K8S_POD_INFRA_CONTAINER_ID":
					env.ParsedArgs.ContainerID = kv[1]
				case "POD_CIDR", "POD_CIDRS", "IPAM_SUBNET":
					env.ParsedArgs.PodCIDR = kv[1]
				}
			}
		}
	}

	return env, nil
}

// CNIResult represents the standard CNI Specification v1.0.0 output format.
type CNIResult struct {
	CNIVersion string             `json:"cniVersion"`
	Interfaces []ResultInterface  `json:"interfaces,omitempty"`
	IPs        []ResultIPConfig   `json:"ips,omitempty"`
	Routes     []ResultRoute      `json:"routes,omitempty"`
	DNS        ResultDNS          `json:"dns,omitempty"`
}

// ResultInterface describes a network interface created by the CNI.
type ResultInterface struct {
	Name    string `json:"name"`
	Mac     string `json:"mac,omitempty"`
	Sandbox string `json:"sandbox,omitempty"` // empty for host interface, netns path for container
}

// ResultIPConfig describes an assigned IP address.
type ResultIPConfig struct {
	Interface *int   `json:"interface,omitempty"` // index into interfaces slice
	Address   string `json:"address"`             // CIDR format, e.g. "10.244.1.5/24"
	Gateway   string `json:"gateway,omitempty"`
}

// ResultRoute describes a routing entry configured by the CNI.
type ResultRoute struct {
	Dst string `json:"dst"`
	GW  string `json:"gw,omitempty"`
}

// ResultDNS describes DNS settings.
type ResultDNS struct {
	Nameservers []string `json:"nameservers,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
}

// HostInterfaceName generates a deterministic host-side interface name from container ID.
func HostInterfaceName(containerID string) string {
	if len(containerID) > 11 {
		return "sg-" + containerID[:11]
	}
	return "sg-" + containerID
}

// WriteResult writes JSON-encoded CNIResult to stdout.
func WriteResult(res *CNIResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// ResolveSocketPath resolves the StraitD API socket path from NetConf, environment, or default.
func ResolveSocketPath(netConfSocket string) string {
	if netConfSocket != "" {
		return netConfSocket
	}
	if env := os.Getenv("STRAITD_API_SOCKET"); env != "" {
		return env
	}
	return "unix:///run/straitd/api.sock"
}

