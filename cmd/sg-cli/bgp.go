// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBgpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bgp",
		Short: "BGP peering and BFD session management",
		Long:  `Manage BGP route advertisements, peer status, and BFD fast-failure detection.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "peers",
		Short: "List configured BGP peers and session status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("PEER IP\t\tREMOTE ASN\tLOCAL ASN\tSTATE\t\tUPTIME\tBFD")
			fmt.Println("10.0.0.1\t65001\t\t65000\t\tEstablished\t4h12m\tUp")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "routes",
		Short: "List routes advertised and received via BGP",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("PREFIX\t\tNEXTHOP\t\tORIGIN\tAS-PATH\t\tSTATUS")
			fmt.Println("10.244.0.0/16\t192.168.1.1\tIGP\t65000\t\tAdvertised")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "bfd",
		Short: "List BFD monitoring sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("PEER\t\tSTATE\tTX-INTERVAL\tRX-INTERVAL\tDETECT-MULT")
			fmt.Println("10.0.0.1\tUp\t300ms\t\t300ms\t\t3")
			return nil
		},
	})

	return cmd
}
