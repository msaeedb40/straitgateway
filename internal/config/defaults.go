package config

import "time"

// Defaults returns a Config populated with production-safe defaults.
func Defaults() Config {
	return Config{
		Datapath: DatapathConfig{
			Mode:           "netkit",
			EnableXDP:      true,
			EnableTCX:      true,
			EnableSocketLB: true,
			EnableLSM:      false, // opt-in
			BPFDir:         "/sys/fs/bpf/straitgateway",
		},
		CNI: CNIConfig{
			ConfDir:  "/etc/cni/net.d",
			BinDir:   "/opt/cni/bin",
			MTU:      0, // auto-detect
			LogLevel: "info",
		},
		IPAM: IPAMConfig{
			Mode: "cluster-pool",
		},
		Service: ServiceConfig{
			Enabled:               true,
			EnableNodePort:        true,
			EnableExternalIPs:     true,
			EnableSessionAffinity: true,
			EnableDSR:             false,
		},
		Policy: PolicyConfig{
			Enabled:         true,
			EnableAuditMode: false,
		},
		Transit: TransitConfig{
			Enabled:     false,
			Topology:    "hub-spoke",
			StraitDAPI:  "unix:///run/straitd/api.sock",
			PeerTimeout: 30 * time.Second,
		},
		Mesh: MeshConfig{
			Enabled: false,
		},
		Observability: ObservabilityConfig{
			EnablePrometheus:    true,
			PrometheusAddr:      ":9090",
			EnableOpenTelemetry: false,
			LogLevel:            "info",
			LogFormat:           "json",
		},
		API: APIConfig{
			Address: "unix:///run/straitd/api.sock",
		},
	}
}
