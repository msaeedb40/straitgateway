#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Build complete distribution bundle in dist/
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"

DIST_DIR="${REPO_ROOT}/dist"
VERSION="${VERSION:-1.0.0}"
export VERSION="${VERSION#v}"

echo "=========================================================="
echo "==> Building Straitgateway v${VERSION} Complete Distribution"
echo "=========================================================="

mkdir -p "${DIST_DIR}/charts" \
         "${DIST_DIR}/crds" \
         "${DIST_DIR}/releases" \
         "${DIST_DIR}/bin" \
         "${DIST_DIR}/ui"

# Step 1: Generate & Bundle CRDs
echo "==> [1/5] Generating CRDs and standalone bundle..."
bash "${REPO_ROOT}/hack/generate/generate-crds.sh"

# Step 2: Build Multi-Arch Binaries & Release Archives
echo "==> [2/5] Building multi-arch binaries and release tarballs..."
bash "${REPO_ROOT}/hack/build/build-all.sh"

# Step 3: Package & Index Helm Charts
echo "==> [3/5] Packaging and indexing Helm charts..."
bash "${REPO_ROOT}/hack/pages/generate-pages.sh"

# Step 4: Build & Stage Angular 22 Dashboard UI
echo "==> [4/5] Building and staging Angular 22 UI..."
if command -v npm &>/dev/null && [[ -d "${REPO_ROOT}/ui" ]]; then
  if [[ ! -d "${REPO_ROOT}/ui/node_modules" ]]; then
    echo "  → Installing UI dependencies (npm ci)..."
    npm --prefix "${REPO_ROOT}/ui" ci --prefer-offline || npm --prefix "${REPO_ROOT}/ui" install
  fi
  echo "  → Compiling production UI..."
  npm --prefix "${REPO_ROOT}/ui" run build
  if [[ -d "${REPO_ROOT}/ui/dist/ui/browser" ]]; then
    cp -r "${REPO_ROOT}/ui/dist/ui/browser/"* "${DIST_DIR}/ui/"
  elif [[ -d "${REPO_ROOT}/ui/dist/ui" ]]; then
    cp -r "${REPO_ROOT}/ui/dist/ui/"* "${DIST_DIR}/ui/"
  fi
  echo "  ✓ UI staged in ${DIST_DIR}/ui/"
else
  echo "  WARN: npm not available or ui/ missing, skipping UI compilation"
fi

# Step 5: Summary & Verification
echo "==> [5/5] Distribution manifest verification..."
echo "----------------------------------------------------------"
echo "Distribution directory layout (${DIST_DIR}):"
tree -L 3 "${DIST_DIR}" 2>/dev/null || find "${DIST_DIR}" -maxdepth 3 -print
echo "----------------------------------------------------------"
echo "✓ Straitgateway distribution build completed successfully!"
