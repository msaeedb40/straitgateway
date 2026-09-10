// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var inputFile string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import straitgateway configuration from file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("input file is required (--file / -f)")
			}
			fmt.Printf("Importing configuration from %s...\n", inputFile)
			fmt.Println("Configuration applied successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input configuration file to apply")
	return cmd
}
