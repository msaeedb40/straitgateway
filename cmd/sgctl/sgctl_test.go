package main

import (
	"bytes"
	"testing"
)

func TestSGCTLCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"version", []string{"version"}},
		{"health", []string{"health"}},
		{"bpf maps list", []string{"bpf", "maps", "list"}},
		{"routes list", []string{"routes", "list"}},
		{"peers status", []string{"peers", "status"}},
		{"conntrack dump", []string{"conntrack", "dump"}},
		{"policy list", []string{"policy", "list"}},
		{"identity list", []string{"identity", "list"}},
		{"service list", []string{"service", "list"}},
		{"metrics fallback", []string{"metrics", "--endpoint", "http://127.0.0.1:54321"}},
		{"flows fallback", []string{"flows", "--endpoint", "http://127.0.0.1:54321"}},
		{"clustermesh status", []string{"clustermesh", "status"}},
		{"clustermesh services", []string{"clustermesh", "services"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tc.args)

			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("command failed with error: %v", err)
			}
		})
	}
}
