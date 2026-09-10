// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiEndpoint  string
	outputFormat string
	kubeconfig   string
	namespace    string
)

var rootCmd = &cobra.Command{
	Use:   "sg-cli",
	Short: "straitgateway CLI — eBPF-native Kubernetes networking & transit gateway",
	Long: `sg-cli is the command-line interface for managing straitgateway.

straitgateway is an eBPF-native CNI, complete kube-proxy replacement,
L3/L4 NetworkPolicy engine, Gateway API v1.6.1 controller, multi-cluster
transit gateway, and BGP/BFD dynamic routing platform.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiEndpoint, "api", "http://localhost:8080", "straitgateway controller API endpoint")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "straitgateway-system", "Kubernetes namespace for straitgateway")

	// Register all modular subcommands
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newGatewayCmd())
	rootCmd.AddCommand(newNodeCmd())
	rootCmd.AddCommand(newBgpCmd())
	rootCmd.AddCommand(newPolicyCmd())
	rootCmd.AddCommand(newTransitCmd())
	rootCmd.AddCommand(newEndpointCmd())
	rootCmd.AddCommand(newClusterCmd())
	rootCmd.AddCommand(newWireguardCmd())
	rootCmd.AddCommand(newIpsecCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newUiCmd())
	rootCmd.AddCommand(newExportCmd())
	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newInstallCmd())
	rootCmd.AddCommand(newUpgradeCmd())
	rootCmd.AddCommand(newVersionCmd())
}
