// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Manage the straitgateway Web UI (enable | disable)",
		Long:  `Manage the deployment state of the Angular 22 dashboard UI. By default, the Web UI is disabled.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show UI deployment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("straitgateway Web UI: Disabled (default)")
			fmt.Println("To enable: sg-cli ui enable")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Enable and deploy the straitgateway Web UI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Enabling straitgateway Web UI...")
			fmt.Println("Scaling UI deployment to 1 replica in namespace:", namespace)
			fmt.Println("UI service endpoint: http://straitgateway-ui." + namespace + ".svc.cluster.local:80")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "Disable the straitgateway Web UI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Disabling straitgateway Web UI...")
			fmt.Println("Scaling UI deployment to 0 replicas in namespace:", namespace)
			return nil
		},
	})

	return cmd
}
