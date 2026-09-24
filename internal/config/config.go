// Package config holds StraitGateway runtime configuration.
package config

import "time"

// Config is the root configuration for all StraitGateway components.
type Config struct {
	// Node is the name of the local Kubernetes node.
	Node string `yaml:"node"`

	// Datapath configures the eBPF datapath.
	Datapath DatapathConfig `yaml:"datapath"`

	// CNI configures the CNI plugin.
	CNI CNIConfig `yaml:"cni"`

	// IPAM configures IP address management.
	IPAM IPAMConfig `yaml:"ipam"`

	// Service configures the service load balancer.
	Service ServiceConfig `yaml:"service"`

	// Policy configures network policy enforcement.
	Policy PolicyConfig `yaml:"policy"`

	// Transit configures the transit gateway daemon.
	Transit TransitConfig `yaml:"transit"`

	// Mesh configures the sidecarless service mesh.
	Mesh MeshConfig `yaml:"mesh"`

	// Observability configures metrics and telemetry.
	Observability ObservabilityConfig `yaml:"observability"`

	// API configures the straitd internal API server.
	API APIConfig `yaml:"api"`
}

// DatapathConfig defines eBPF datapath settings.
type DatapathConfig struct {
	// Mode: "netkit" (default) or "veth" (fallback).
	Mode string `yaml:"mode"`
	// EnableXDP enables XDP attachment on uplink interfaces.
	EnableXDP bool `yaml:"enableXDP"`
	// EnableTCX enables TCX attachment on Netkit interfaces.
	EnableTCX bool `yaml:"enableTCX"`
	// EnableSocketLB enables socket-level load balancing via cgroup hooks.
	EnableSocketLB bool `yaml:"enableSocketLB"`
	// EnableLSM enables LSM-based policy enforcement.
	EnableLSM bool `yaml:"enableLSM"`
	// BPFDir is the directory for pinned eBPF objects.
	BPFDir string `yaml:"bpfDir"`
}

// CNIConfig holds CNI plugin settings.
type CNIConfig struct {
	ConfDir  string `yaml:"confDir"`  // /etc/cni/net.d
	BinDir   string `yaml:"binDir"`   // /opt/cni/bin
	MTU      int    `yaml:"mtu"`      // 0 = auto-detect
	LogLevel string `yaml:"logLevel"` // debug|info|warn|error
}

// IPAMConfig holds IPAM settings.
type IPAMConfig struct {
	// Mode: "cluster-pool" or "multi-pool".
	Mode     string   `yaml:"mode"`
	PodCIDRs []string `yaml:"podCIDRs"`
}

// ServiceConfig holds kube-proxy replacement settings.
type ServiceConfig struct {
	Enabled               bool `yaml:"enabled"`
	EnableNodePort        bool `yaml:"enableNodePort"`
	EnableExternalIPs     bool `yaml:"enableExternalIPs"`
	EnableSessionAffinity bool `yaml:"enableSessionAffinity"`
	EnableDSR             bool `yaml:"enableDSR"` // Direct Server Return
}

// PolicyConfig holds policy enforcement settings.
type PolicyConfig struct {
	Enabled          bool `yaml:"enabled"`
	EnableAuditMode  bool `yaml:"enableAuditMode"` // log-only, no drops
}

// TransitConfig holds transit gateway settings.
type TransitConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Topology   string        `yaml:"topology"` // hub-spoke|mesh|peer-to-peer|hub-to-hub|hybrid
	StraitDAPI string        `yaml:"straitdAPI"` // address of straitd internal API
	PeerTimeout time.Duration `yaml:"peerTimeout"`
}

// MeshConfig holds sidecarless mesh settings.
type MeshConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ObservabilityConfig holds observability settings.
type ObservabilityConfig struct {
	EnablePrometheus    bool   `yaml:"enablePrometheus"`
	PrometheusAddr      string `yaml:"prometheusAddr"` // :9090
	EnableOpenTelemetry bool   `yaml:"enableOpenTelemetry"`
	OTLPEndpoint        string `yaml:"otlpEndpoint"` // grpc endpoint
	LogLevel            string `yaml:"logLevel"`     // debug|info|warn|error
	LogFormat           string `yaml:"logFormat"`    // json|text
}

// APIConfig holds the straitd internal gRPC API server settings.
type APIConfig struct {
	// Address is the listen address for the internal API.
	// Used by tgwd and sgctl to communicate with straitd.
	Address string `yaml:"address"` // unix:///run/straitd/api.sock
}
