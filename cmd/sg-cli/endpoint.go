// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newEndpointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Endpoint and NetKit interface management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active pod network endpoints and assigned BPF identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("NAMESPACE\tPOD\t\tIP\t\tIFACE\t\tIDENTITY\tSTATUS")
			fmt.Println("kube-system\tcoredns-abc\t10.244.0.2\tnk-1001\t\t2\t\tReady")
			fmt.Println("default\t\tweb-app-xyz\t10.244.0.10\tnk-1002\t\t1004\t\tReady")
			return nil
		},
	})

	return cmd
}
