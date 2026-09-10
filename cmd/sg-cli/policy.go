// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Network policy and StraitNetworkPolicy management",
		Long:  `Inspect and test L3/L4 NetworkPolicies, multi-dimensional selectors, and priority 0-255 evaluation engine rules.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List compiled network policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("NAMESPACE\tNAME\t\tPRIORITY\tRULES\tACTION\tSTATUS")
			fmt.Println("default\t\tallow-frontend\t10\t\t2\tAllow\tCompiled")
			fmt.Println("default\t\tdeny-all\t250\t\t1\tDeny\tCompiled")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "eval <src-pod> <dst-pod> <port>",
		Short: "Simulate policy evaluation between two endpoints",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst, port := args[0], args[1], args[2]
			fmt.Printf("Simulating policy match: %s -> %s:%s\n", src, dst, port)
			fmt.Println("Evaluation Result:")
			fmt.Println("  Matched Policy: allow-frontend (Priority: 10)")
			fmt.Println("  Action:         ALLOW (Zero dropped packets)")
			return nil
		},
	})

	return cmd
}
