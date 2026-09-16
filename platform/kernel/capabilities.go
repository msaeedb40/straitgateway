// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package kernel verifies Linux kernel capabilities and prerequisites.
package kernel

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// RequiredCapabilities lists the minimum Linux capabilities required by straitgatewayd:
// NET_ADMIN, SYS_ADMIN, NET_RAW, PERFMON, BPF.
var RequiredCapabilities = []string{
	"CAP_NET_ADMIN",
	"CAP_SYS_ADMIN",
	"CAP_NET_RAW",
	"CAP_PERFMON",
	"CAP_BPF",
}

// CheckPrivileges verifies whether the current process is running as root (UID 0).
func CheckPrivileges(log *zap.Logger) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("straitgatewayd must run as root or with required capabilities (%v)", RequiredCapabilities)
	}
	log.Info("process privilege verified (running as root)")
	return nil
}
