#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
echo "→ Checking gofmt..."
UNFORMATTED=$(gofmt -l .)
if [[ -n "${UNFORMATTED}" ]]; then
  echo "ERROR: Unformatted Go files found:"
  echo "${UNFORMATTED}"
  echo "Run 'gofmt -s -w .' to fix."
  exit 1
fi
echo "✓ All Go files are properly formatted"
