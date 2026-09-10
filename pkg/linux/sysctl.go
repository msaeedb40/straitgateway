// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package linux provides Linux kernel interaction utilities.
package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sysctl reads or writes a sysctl parameter.
type Sysctl struct {
	basePath string
}

// NewSysctl creates a new Sysctl with the default /proc/sys base path.
func NewSysctl() *Sysctl {
	return &Sysctl{basePath: "/proc/sys"}
}

// Get reads a sysctl value.
func (s *Sysctl) Get(key string) (string, error) {
	path := filepath.Join(s.basePath, strings.ReplaceAll(key, ".", "/"))
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading sysctl %s: %w", key, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Set writes a sysctl value.
func (s *Sysctl) Set(key, value string) error {
	path := filepath.Join(s.basePath, strings.ReplaceAll(key, ".", "/"))
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return fmt.Errorf("writing sysctl %s=%s: %w", key, value, err)
	}
	return nil
}

// EnsureForwarding enables IPv4 and IPv6 forwarding.
func (s *Sysctl) EnsureForwarding() error {
	if err := s.Set("net.ipv4.ip_forward", "1"); err != nil {
		return err
	}
	// IPv6 forwarding is best-effort (may not be available).
	_ = s.Set("net.ipv6.conf.all.forwarding", "1")
	return nil
}

// DisableRPFilter disables reverse path filtering for the given interface.
func (s *Sysctl) DisableRPFilter(iface string) error {
	return s.Set(fmt.Sprintf("net.ipv4.conf.%s.rp_filter", iface), "0")
}

// DisableSendRedirects disables ICMP redirect sending for the given interface.
func (s *Sysctl) DisableSendRedirects(iface string) error {
	return s.Set(fmt.Sprintf("net.ipv4.conf.%s.send_redirects", iface), "0")
}
