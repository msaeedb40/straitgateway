// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package resource manages cgroup v2 resource limits (CPU, memory, IO).
package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
)

const cgroupV2Root = "/sys/fs/cgroup"

// Manager manages cgroup v2 limits for straitgateway processes.
type Manager struct {
	log       *zap.Logger
	slicePath string
}

// New creates a new cgroup v2 resource Manager.
func New(sliceName string, log *zap.Logger) *Manager {
	return &Manager{
		log:       log,
		slicePath: filepath.Join(cgroupV2Root, sliceName),
	}
}

// SetCPULimit sets the cpu.max limit (quota in microseconds, period in microseconds).
func (m *Manager) SetCPULimit(quotaUs, periodUs int64) error {
	val := fmt.Sprintf("%d %d\n", quotaUs, periodUs)
	return m.writeControlFile("cpu.max", val)
}

// SetMemoryMax sets the memory.max limit in bytes.
func (m *Manager) SetMemoryMax(bytes int64) error {
	return m.writeControlFile("memory.max", strconv.FormatInt(bytes, 10)+"\n")
}

// SetIOWeight sets the io.weight (1-10000).
func (m *Manager) SetIOWeight(weight int) error {
	return m.writeControlFile("io.weight", strconv.Itoa(weight)+"\n")
}

func (m *Manager) writeControlFile(name, content string) error {
	filePath := filepath.Join(m.slicePath, name)
	if _, err := os.Stat(m.slicePath); os.IsNotExist(err) {
		if err := os.MkdirAll(m.slicePath, 0755); err != nil {
			return fmt.Errorf("creating cgroup dir: %w", err)
		}
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		m.log.Warn("cgroup v2 write failed (may lack permissions or controller)",
			zap.String("file", filePath),
			zap.Error(err),
		)
		return err
	}
	return nil
}
