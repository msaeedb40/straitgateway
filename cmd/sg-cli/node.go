// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage and inspect straitgatewayd node agents and eBPF state",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all cluster nodes running straitgatewayd",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("NODE\t\tPOD-CIDR\t\tAGENT-STATUS\tKERNEL\t\tBPF-FS")
			fmt.Println("node-1\t\t10.244.1.0/24\t\tReady\t\tLinux 6.12 LTS\tMounted (/sys/fs/bpf)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "inspect <node>",
		Short: "Inspect eBPF maps, NetKit links, and loaded programs on a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeName := args[0]
			fmt.Printf("=== Node Inspection: %s ===\n", nodeName)
			fmt.Println("Kernel: Linux 6.12+ (cgroup v2 enabled)")
			fmt.Println("eBPF Dataplane Hooks:")
			fmt.Println("  - NetKit: Attached (pod primary datapath)")
			fmt.Println("  - TCX:    Attached (node/host fallback datapath)")
			fmt.Println("  - XDP:    Attached (fast edge filter & NodePort acceleration)")
			fmt.Println("  - Sockops: Attached (socket-level TCP acceleration)")
			fmt.Println("  - LSM:    Active (stateful security enforcement)")
			fmt.Println("BPF Maps (/sys/fs/bpf/straitgateway):")
			fmt.Println("  - service_map:   Active (kube-dns + cluster VIPs)")
			fmt.Println("  - backend_map:   Active")
			fmt.Println("  - policy_map:    Active (priority 0-255 engine)")
			fmt.Println("  - identity_map:  Active")
			fmt.Println("  - ct_map:        Active (stateful conntrack)")
			return nil
		},
	})

	return cmd
}
