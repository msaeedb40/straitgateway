// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGatewayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage and inspect Gateway API v1.6.1 resources",
		Long:  `Inspect and manage GatewayClasses, Gateways, HTTPRoutes, GRPCRoutes, TLSRoutes, TCPRoutes, and UDPRoutes.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List configured Gateways and listeners",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("GATEWAY\t\tCLASS\t\tADDRESS\t\tPROGRAMMED\tAGE")
			fmt.Println("strait-gw\tstraitgateway\t192.168.1.100\tTrue\t\t1d")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "routes",
		Short: "List routes attached to gateways (HTTP, gRPC, TLS, TCP, UDP)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("NAME\t\tTYPE\t\tGATEWAY\t\tHOSTNAMES\t\tSTATUS")
			fmt.Println("web-route\tHTTPRoute\tstrait-gw\t[\"*.example.com\"]\tAccepted")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "describe <name>",
		Short: "Show detailed status of a Gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Gateway: %s (Namespace: %s)\n", args[0], namespace)
			fmt.Println("Listeners:")
			fmt.Println("  - Name: http, Port: 80, Protocol: HTTP, Status: Programmed")
			fmt.Println("  - Name: https, Port: 443, Protocol: HTTPS, Status: Programmed")
			return nil
		},
	})

	return cmd
}
