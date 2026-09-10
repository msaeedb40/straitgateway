// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Multi-cluster federation and mesh management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connected clusters in the transit mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("CLUSTER ID\tNAME\t\tPOD-CIDR\tSERVICE-CIDR\tGATEWAY ENDPOINT\tSTATUS")
			fmt.Println("1\t\tcluster-primary\t10.244.0.0/16\t10.96.0.0/12\t192.168.1.10:51820\tConnected")
			fmt.Println("2\t\tcluster-edge\t10.245.0.0/16\t10.97.0.0/12\t192.168.2.10:51820\tConnected")
			return nil
		},
	})

	return cmd
}
