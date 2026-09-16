#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHART_DIR="${REPO_ROOT}/straitgateway-helm"
echo "→ Verifying Helm chart templates dry-run..."
if command -v helm &>/dev/null; then
  helm template straitgateway "${CHART_DIR}" --dry-run > /dev/null
  echo "✓ Helm manifests verified"
else
  echo "WARN: helm CLI not installed, skipping helm template check"
fi
