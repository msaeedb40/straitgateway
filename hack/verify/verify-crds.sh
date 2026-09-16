#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CRD_DIR="${REPO_ROOT}/straitgateway-helm/charts/crds/templates"
echo "→ Verifying CRD manifests in ${CRD_DIR}..."
if [[ ! -d "${CRD_DIR}" ]]; then
  echo "ERROR: CRD template directory ${CRD_DIR} not found"
  exit 1
fi
COUNT=$(find "${CRD_DIR}" -name "*.yaml" | wc -l)
if [[ "${COUNT}" -eq 0 ]]; then
  echo "ERROR: No CRDs found in ${CRD_DIR}"
  exit 1
fi
echo "✓ Verified ${COUNT} CRD definitions"
