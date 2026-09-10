#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Verify go modules are tidy and up-to-date.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"

cp go.sum go.sum.bak
go mod tidy
if ! diff -q go.sum go.sum.bak &>/dev/null; then
  echo "ERROR: go.sum is out of date. Run 'go mod tidy' and commit."
  rm go.sum.bak; exit 1
fi
rm go.sum.bak
echo "✓ go modules are tidy"
