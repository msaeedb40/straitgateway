// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTransitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transit",
		Short: "Multi-cluster transit gateway management",
		Long:  `Manage transit segments (32-bit IDs, Segment 0 backbone), attachments, cross-cluster peering, and routes.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "segments",
		Short: "List transit segments",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("SEGMENT-ID\tNAME\t\tROLE\t\tATTACHMENTS\tSTATUS")
			fmt.Println("0\t\tbackbone\tBackbone\t4\t\tActive")
			fmt.Println("1\t\tprod-mesh\tTenant\t\t2\t\tActive")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "attachments",
		Short: "List transit segment attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ATTACHMENT\t\tSEGMENT\tCLUSTER\t\tTYPE\t\tSTATUS")
			fmt.Println("TransitAttachment1\t0\tcluster-primary\tVPC/Cluster\tAttached")
			fmt.Println("TransitAttachment2\t1\tcluster-edge\tVPC/Cluster\tAttached")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "routes",
		Short: "List transit segment routes",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("SEGMENT\tCIDR\t\tNEXTHOP\t\t\tSTATUS")
			fmt.Println("0\t0.0.0.0/0\tTransitAttachment1\tActive")
			fmt.Println("1\t10.100.0.0/16\tTransitAttachment2\tActive")
			return nil
		},
	})

	return cmd
}
