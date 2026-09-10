// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/msaeedb40/straitgateway/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print straitgateway version, build metadata, and kernel support baseline",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("straitgateway CLI:")
			fmt.Printf("  Version:           %s\n", version.Version)
			fmt.Printf("  Git Commit:        %s\n", version.Commit)
			fmt.Printf("  Build Date:        %s\n", version.BuildDate)
			fmt.Printf("  Go Version:        %s\n", version.GoVersion)
			fmt.Println("Kernel Support:")
			fmt.Println("  Minimum Kernel:    Linux 6.7")
			fmt.Println("  Production LTS:    Linux 6.12 LTS")
			fmt.Println("  Supported Arch:    amd64, arm64")
			fmt.Println("  Features:          NetKit, TCX, XDP, LSM, cgroup v2, Gateway API v1.6.1")
		},
	}
}
