#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Generate Go stubs from protobuf definitions using buf.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}/proto"
if ! command -v buf &>/dev/null; then
  echo "ERROR: 'buf' not found. Install from https://buf.build/docs/installation"
  exit 1
fi
echo "→ Generating proto stubs..."
buf generate
echo "✓ Proto stubs generated"
