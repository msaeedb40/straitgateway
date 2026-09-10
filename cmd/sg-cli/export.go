// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export straitgateway topology and network policy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputFile != "" {
				fmt.Printf("Exporting configuration to %s...\n", outputFile)
			} else {
				fmt.Println("# Exported straitgateway configuration (YAML)")
				fmt.Println("apiVersion: networking.straitgateway.io/v1alpha1")
				fmt.Println("kind: ClusterNetworkConfig")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "file", "f", "", "Output file destination")
	return cmd
}
