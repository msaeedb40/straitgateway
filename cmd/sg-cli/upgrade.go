// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	var (
		chartVersion string
		valuesFile   string
	)

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade straitgateway installation via Helm",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("=== Upgrading straitgateway ===")
			fmt.Printf("Namespace: %s\n", namespace)
			fmt.Printf("helm upgrade straitgateway straitgateway/straitgateway --namespace %s\n", namespace)
			return nil
		},
	}

	cmd.Flags().StringVar(&chartVersion, "version", "", "Target chart version for upgrade")
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "Path to custom values.yaml file")
	return cmd
}
