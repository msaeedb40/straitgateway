#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}/bpf"
echo "→ Verifying BPF C headers and map schemas..."
if [[ ! -f "maps/maps.h" ]]; then
  echo "ERROR: maps/maps.h missing"
  exit 1
fi
echo "✓ BPF schemas present"
