#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
if command -v k3s-uninstall.sh &>/dev/null; then
  k3s-uninstall.sh
elif command -v k3s-agent-uninstall.sh &>/dev/null; then
  k3s-agent-uninstall.sh
fi
echo "[straitgateway] k3s uninstalled."
