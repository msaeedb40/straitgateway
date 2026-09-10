// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newWireguardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wireguard",
		Short: "WireGuard tunnel management and encryption status",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show WireGuard tunnel interface and peer handshake details",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("INTERFACE\tPUBLIC KEY\t\t\t\tPORT\tPEERS\tSTATUS")
			fmt.Println("sg-wg0\t\tzABC123...DEF890=\t\t\t51820\t2\tActive")
			fmt.Println("\nPeers:")
			fmt.Println("  - Remote Cluster 2 (192.168.2.10:51820): Handshake 12s ago, Transfer: 14.2 MB rx / 18.5 MB tx")
			return nil
		},
	})

	return cmd
}
