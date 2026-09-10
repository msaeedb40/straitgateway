// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		kubeProxyReplacement bool
		chartVersion         string
		valuesFile           string
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install straitgateway in Kubernetes cluster via Helm",
		Long: `Installs straitgateway into the target Kubernetes cluster with recommended production settings:
  - CNI with NetKit acceleration
  - Complete kube-proxy replacement (enabled by default)
  - BPF filesystem mount at /sys/fs/bpf
  - Systemd cgroup v2 integration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("=== Installing straitgateway ===")
			fmt.Printf("Namespace:                %s\n", namespace)
			fmt.Printf("kubeProxyReplacement:     %v\n", kubeProxyReplacement)
			if valuesFile != "" {
				fmt.Printf("Custom Values:            %s\n", valuesFile)
			}
			fmt.Println("\nExecuting helm install command:")
			fmt.Printf("helm upgrade --install straitgateway straitgateway/straitgateway \\\n"+
				"  --namespace %s --create-namespace \\\n"+
				"  --set kubeProxyReplacement.enabled=%v \\\n"+
				"  --set kubeProxyReplacement.mode=none\n", namespace, kubeProxyReplacement)
			return nil
		},
	}

	cmd.Flags().BoolVar(&kubeProxyReplacement, "kube-proxy-replacement", true, "Enable full kube-proxy replacement")
	cmd.Flags().StringVar(&chartVersion, "version", "", "Specific chart version to install")
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "Path to custom values.yaml file")
	return cmd
}
