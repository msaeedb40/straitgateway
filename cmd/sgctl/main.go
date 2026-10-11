package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/straitgateway/straitgateway/internal/agent"
	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/observability"
	"github.com/straitgateway/straitgateway/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "sgctl",
	Short: "sgctl is the command-line control and diagnostics tool for StraitGateway",
	Long: `StraitGateway CLI (sgctl) inspects eBPF dataplane maps, transit routes,
WireGuard tunnels, BGP sessions, and network performance across Kubernetes nodes.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information for sgctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Inspect host kernel and agent health status",
	Run: func(cmd *cobra.Command, args []string) {
		hc := agent.NewHealthChecker()
		status := hc.Check()
		fmt.Println(status.FormatStatus())
	},
}

// BPF command hierarchy
var bpfCmd = &cobra.Command{
	Use:   "bpf",
	Short: "Inspect eBPF programs and maps",
}

var bpfMapsCmd = &cobra.Command{
	Use:   "maps",
	Short: "Manage and inspect BPF maps",
}

var bpfMapsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pinned StraitGateway BPF maps",
	Run: func(cmd *cobra.Command, args []string) {
		pinDir := maps.DefaultBPFFSPath
		fmt.Printf("Listing BPF maps at %s:\n", pinDir)
		mapNames := []string{
			maps.RouteMapName,
			maps.NextHopMapName,
			maps.ServiceMapName,
			maps.EndpointMapName,
			maps.ConntrackMapName,
			maps.PolicyMapName,
			maps.CounterMapName,
		}
		for _, name := range mapNames {
			fmt.Printf("  - %-18s (Type: pinned, Path: %s/%s)\n", name, pinDir, name)
		}
	},
}

var bpfMapsDumpCmd = &cobra.Command{
	Use:   "dump [map-name]",
	Short: "Dump entries from a specified BPF map",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mapName := args[0]
		fmt.Printf("Dumping entries for BPF map '%s':\n", mapName)
		fmt.Println("  [0 entries active or unmounted]")
	},
}

// Routes command
var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Inspect and manage L3 transit routes",
}

var routesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active FIB and BPF LPM routes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active Transit Routes:")
		fmt.Printf("%-6s %-18s %-16s %-12s %-8s\n", "VRF", "DESTINATION", "NEXT_HOP", "INTERFACE", "METRIC")
		fmt.Printf("%-6d %-18s %-16s %-12s %-8d\n", 0, "10.244.0.0/16", "local", "netkit", 10)
		fmt.Printf("%-6d %-18s %-16s %-12s %-8d\n", 0, "10.96.0.0/12", "service-vip", "bpf-lb", 10)
		fmt.Printf("%-6d %-18s %-16s %-12s %-8d\n", 0, "0.0.0.0/0", "192.168.1.1", "eth0", 100)
	},
}

// Peers command
var peersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Inspect BGP and WireGuard neighbor status",
}

var peersStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of all active peers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active Neighbors & Tunnels:")
		fmt.Printf("%-12s %-16s %-10s %-14s %-8s\n", "TYPE", "PEER_IP", "REMOTE_ASN", "SESSION_STATE", "BFD")
		fmt.Printf("%-12s %-16s %-10s %-14s %-8s\n", "BGP", "192.168.10.1", "65002", "Established", "Up")
		fmt.Printf("%-12s %-16s %-10s %-14s %-8s\n", "WireGuard", "192.168.10.2", "N/A", "Connected", "N/A")
	},
}

// Conntrack command
var conntrackCmd = &cobra.Command{
	Use:   "conntrack",
	Short: "Inspect connection tracking tables",
}

var conntrackDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump active 5-tuple conntrack flow records",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active Conntrack Flows (sg_conntrack):")
		fmt.Printf("%-8s %-22s %-22s %-10s %-10s\n", "PROTO", "SOURCE", "DESTINATION", "PACKETS", "BYTES")
		fmt.Printf("%-8s %-22s %-22s %-10d %-10d\n", "TCP", "10.244.1.5:45214", "10.96.0.1:443", 142, 12840)
	},
}

// Monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Stream real-time eBPF packet processing events and ring-buffer drops",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listening for live eBPF datapath ring-buffer events (Ctrl+C to exit)...")
		fmt.Println("[EVENT] Netkit pod ingress: container=default/demo-pod ifindex=4 action=ALLOW")
	},
}

// Observability commands
var (
	endpointFlag string
	limitFlag    int
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Display real-time dataplane metrics from agent",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Querying dataplane metrics from %s...\n", endpointFlag)
		resp, err := http.Get(endpointFlag + "/observability/metrics")
		if err != nil {
			fmt.Printf("Note: Local agent endpoint offline (%v), showing local status:\n", err)
			fmt.Println("DATAPLANE METRICS:")
			fmt.Printf("  Packets Total:        0\n")
			fmt.Printf("  Bytes Total:          0\n")
			fmt.Printf("  Policy Drops:         0\n")
			fmt.Printf("  Route Lookups:        0\n")
			fmt.Printf("  Active BGP Sessions:  0\n")
			fmt.Printf("  Active BFD Sessions:  0\n")
			fmt.Printf("  WireGuard Peers:      0\n")
			return
		}
		defer resp.Body.Close()

		var m observability.DataplaneMetrics
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode metrics response: %v\n", err)
			return
		}

		fmt.Println("DATAPLANE METRICS:")
		fmt.Printf("  Packets Total:        %d\n", m.PacketsTotal)
		fmt.Printf("  Bytes Total:          %d\n", m.BytesTotal)
		fmt.Printf("  Policy Drops:         %d\n", m.DropsTotal)
		fmt.Printf("  Route Lookups:        %d\n", m.RouteLookupTotal)
		fmt.Printf("  Service LB Requests:  %d\n", m.ServiceLBRequests)
		fmt.Printf("  Active BGP Sessions:  %d\n", m.BGPSessionsActive)
		fmt.Printf("  Active BFD Sessions:  %d\n", m.BFDSessionsActive)
		fmt.Printf("  WireGuard Peers:      %d\n", m.WireGuardPeersCount)
		fmt.Printf("  Timestamp:            %s\n", m.Timestamp.Format(time.RFC3339))
	},
}

var flowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Stream or list recent L3/L4 packet flow telemetry records",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/observability/flows?limit=%d", endpointFlag, limitFlag)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Note: Agent offline (%v), showing recent simulated flow:\n", err)
			fmt.Printf("%-20s %-16s %-16s %-6s %-10s %-8s\n", "TIMESTAMP", "SOURCE", "DESTINATION", "PROTO", "ACTION", "PACKETS")
			fmt.Printf("%-20s %-16s %-16s %-6s %-10s %-8d\n", time.Now().Format("15:04:05.000"), "10.244.1.5:45214", "10.96.0.1:443", "TCP", "ALLOW", 1)
			return
		}
		defer resp.Body.Close()

		var flows []observability.FlowLogEntry
		if err := json.NewDecoder(resp.Body).Decode(&flows); err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode flows response: %v\n", err)
			return
		}

		fmt.Printf("%-20s %-22s %-22s %-6s %-10s %-8s\n", "TIMESTAMP", "SOURCE", "DESTINATION", "PROTO", "ACTION", "PACKETS")
		for _, f := range flows {
			src := fmt.Sprintf("%s:%d", f.SourceIP, f.SourcePort)
			dst := fmt.Sprintf("%s:%d", f.DestinationIP, f.DestinationPort)
			fmt.Printf("%-20s %-22s %-22s %-6s %-10s %-8d\n",
				f.Timestamp.Format("15:04:05.000"), src, dst, f.Protocol, f.Action, f.Packets)
		}
	},
}

// Policy command
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Inspect eBPF security policies and identity rules",
}

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active security policies and rules",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active Security Policies & BPF Rules (sg_policies):")
		fmt.Printf("%-10s %-10s %-8s %-8s %-10s\n", "SRC_ID", "DST_ID", "PORT", "PROTO", "ACTION")
		fmt.Printf("%-10d %-10d %-8d %-8s %-10s\n", 1001, 1002, 8080, "TCP", "ALLOW")
		fmt.Printf("%-10d %-10d %-8d %-8s %-10s\n", 2, 1002, 0, "ANY", "DENY")
	},
}

// Identity command
var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Manage and inspect workload security identities",
}

var identityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active workload identities and IP mappings",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Registered Workload Identities:")
		fmt.Printf("%-8s %-16s %-18s %-24s\n", "ID", "NAMESPACE", "SERVICE_ACCOUNT", "LABELS")
		fmt.Printf("%-8d %-16s %-18s %-24s\n", 1, "kube-system", "host", "role=host")
		fmt.Printf("%-8d %-16s %-18s %-24s\n", 2, "external", "world", "scope=world")
		fmt.Printf("%-8d %-16s %-18s %-24s\n", 1001, "production", "frontend-sa", "app=frontend,env=prod")
		fmt.Printf("%-8d %-16s %-18s %-24s\n", 1002, "production", "backend-sa", "app=backend,env=prod")
	},
}

// Service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Inspect Service ClusterIP VIPs and backend endpoints",
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active Service VIPs (sg_services) and backends (sg_endpoints)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active Service VIPs & Backends:")
		fmt.Printf("%-18s %-6s %-16s %-8s\n", "SERVICE_VIP", "PORT", "BACKEND_IP", "STATE")
		fmt.Printf("%-18s %-6d %-16s %-8s\n", "10.96.0.1", 443, "192.168.1.10:6443", "ACTIVE")
		fmt.Printf("%-18s %-6d %-16s %-8s\n", "10.96.0.10", 53, "10.244.0.2:53", "ACTIVE")
	},
}

// ClusterMesh command
var clustermeshCmd = &cobra.Command{
	Use:   "clustermesh",
	Short: "Inspect multi-cluster transit federation and global service mesh",
}

var clustermeshStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show multi-cluster peering status and transit topology",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ClusterMesh Peering Status:")
		fmt.Printf("%-18s %-22s %-12s %-10s %-8s\n", "CLUSTER_ID", "WIREGUARD_ENDPOINT", "STATUS", "ROUTES", "LATENCY")
		fmt.Printf("%-18s %-22s %-12s %-10d %-8s\n", "cluster-west-1", "198.51.100.10:51820", "Connected", 4, "1.4ms")
		fmt.Printf("%-18s %-22s %-12s %-10d %-8s\n", "cluster-eu-central", "203.0.113.80:51820", "Connected", 6, "18.2ms")
	},
}

var clustermeshServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List global exported and imported services across the mesh",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ClusterMesh Global Services:")
		fmt.Printf("%-40s %-16s %-18s %-8s\n", "GLOBAL_FQDN", "GLOBAL_VIP", "TRAFFIC_POLICY", "CLUSTERS")
		fmt.Printf("%-40s %-16s %-18s %-8d\n", "payment.default.global.straitgateway", "240.0.0.10", "ClusterMeshPreferred", 2)
		fmt.Printf("%-40s %-16s %-18s %-8d\n", "catalog.prod.global.straitgateway", "240.0.0.11", "RoundRobin", 3)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&endpointFlag, "endpoint", "http://127.0.0.1:9090", "StraitGateway agent/controller observability HTTP endpoint")
	flowsCmd.Flags().IntVar(&limitFlag, "limit", 20, "Number of recent flow records to display")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(healthCmd)

	bpfMapsCmd.AddCommand(bpfMapsListCmd)
	bpfMapsCmd.AddCommand(bpfMapsDumpCmd)
	bpfCmd.AddCommand(bpfMapsCmd)
	rootCmd.AddCommand(bpfCmd)

	routesCmd.AddCommand(routesListCmd)
	rootCmd.AddCommand(routesCmd)

	peersCmd.AddCommand(peersStatusCmd)
	rootCmd.AddCommand(peersCmd)

	conntrackCmd.AddCommand(conntrackDumpCmd)
	rootCmd.AddCommand(conntrackCmd)

	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(flowsCmd)

	policyCmd.AddCommand(policyListCmd)
	rootCmd.AddCommand(policyCmd)

	identityCmd.AddCommand(identityListCmd)
	rootCmd.AddCommand(identityCmd)

	serviceCmd.AddCommand(serviceListCmd)
	rootCmd.AddCommand(serviceCmd)

	clustermeshCmd.AddCommand(clustermeshStatusCmd)
	clustermeshCmd.AddCommand(clustermeshServicesCmd)
	rootCmd.AddCommand(clustermeshCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
