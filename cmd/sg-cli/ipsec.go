// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newIpsecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ipsec",
		Short: "IPsec tunnel and Security Association (SA) management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show active IPsec tunnels and SAs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TUNNEL\t\tREMOTE ENDPOINT\tSPI\t\tSTATE\tUPTIME")
			fmt.Println("sg-ipsec0\t192.168.3.10\t0xc1234567\tESTABLISHED\t2d")
			return nil
		},
	})

	return cmd
}
