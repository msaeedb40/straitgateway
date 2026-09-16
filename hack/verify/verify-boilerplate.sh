#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Verify all Go files have the SPDX license header.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
missing=()
while IFS= read -r -d '' f; do
  if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$f"; then
    missing+=("$f")
  fi
done < <(find "${REPO_ROOT}" -name "*.go" \
  -not -path "*/vendor/*" \
  -not -path "*/.git/*" \
  -not -name "*.pb.go" \
  -not -name "*_bpf*.go" \
  -not -name "zz_generated*.go" \
  -print0)
if [ ${#missing[@]} -gt 0 ]; then
  echo "ERROR: Missing SPDX header in ${#missing[@]} file(s):"
  printf '  %s\n' "${missing[@]}"
  exit 1
fi
echo "✓ All Go files have SPDX license headers"
