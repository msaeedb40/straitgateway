#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Verify generated code (DeepCopy, CRDs) is up-to-date.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
make generate
if ! git diff --exit-code; then
  echo "ERROR: Generated files are out of date. Run 'make generate' and commit."
  exit 1
fi
echo "✓ Generated code is up-to-date"
