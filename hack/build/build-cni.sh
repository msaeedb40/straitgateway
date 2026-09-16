#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
echo "→ Building straitgateway CNI plugin binary..."
mkdir -p bin
CGO_ENABLED=0 go build -o bin/straitgateway-cni ./cni
echo "✓ Built bin/straitgateway-cni"
