// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management for straitgateway",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "view",
		Short: "View current straitgateway configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("=== straitgateway Configuration ===")
			fmt.Printf("API Endpoint:            %s\n", apiEndpoint)
			fmt.Printf("Namespace:               %s\n", namespace)
			fmt.Println("kubeProxyReplacement:    true")
			fmt.Println("kubeProxyMode:           \"none\"")
			fmt.Println("BPF Filesystem:          /sys/fs/bpf/straitgateway")
			fmt.Println("Socket Path:             /run/straitgateway/daemon.sock")
			return nil
		},
	})

	return cmd
}
