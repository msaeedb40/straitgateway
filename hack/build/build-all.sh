#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Build multi-architecture binaries and distribution archives for Straitgateway.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"

DIST_DIR="${REPO_ROOT}/dist"
RELEASES_DIR="${DIST_DIR}/releases"
BIN_DIST_DIR="${DIST_DIR}/bin"

mkdir -p "${RELEASES_DIR}" "${BIN_DIST_DIR}"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 1.0.0)}"
VERSION="${VERSION#v}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

LDFLAGS="-s -w \
  -X github.com/msaeedb40/straitgateway/internal/version.Version=${VERSION} \
  -X github.com/msaeedb40/straitgateway/internal/version.Commit=${COMMIT} \
  -X github.com/msaeedb40/straitgateway/internal/version.BuildDate=${BUILD_DATE}"

PLATFORMS=("linux/amd64" "linux/arm64")

echo "=========================================================="
echo "==> Building Straitgateway v${VERSION} Multi-Arch Binaries"
echo "=========================================================="

for platform in "${PLATFORMS[@]}"; do
  GOOS="${platform%/*}"
  GOARCH="${platform#*/}"
  TARGET_NAME="${GOOS}-${GOARCH}"
  TARGET_DIR="${BIN_DIST_DIR}/${TARGET_NAME}"
  STAGE_DIR="${RELEASES_DIR}/stage/${TARGET_NAME}"

  echo "==> Building ${TARGET_NAME} binaries..."
  mkdir -p "${TARGET_DIR}" "${STAGE_DIR}"

  for cmd in straitgatewayd sg-controller sg-cli; do
    echo "  → ${cmd} (${TARGET_NAME})"
    CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
      -ldflags="${LDFLAGS}" \
      -o "${TARGET_DIR}/${cmd}" \
      "./cmd/${cmd}"
    cp "${TARGET_DIR}/${cmd}" "${STAGE_DIR}/${cmd}"
  done

  echo "  → straitgateway-cni (${TARGET_NAME})"
  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -ldflags="${LDFLAGS}" \
    -o "${TARGET_DIR}/straitgateway-cni" \
    "./cni"
  cp "${TARGET_DIR}/straitgateway-cni" "${STAGE_DIR}/straitgateway-cni"

  # Include docs and license
  cp LICENSE README.md "${STAGE_DIR}/"

  TARBALL_NAME="straitgateway-v${VERSION}-${TARGET_NAME}.tar.gz"
  echo "==> Packaging ${TARBALL_NAME}..."
  tar -czf "${RELEASES_DIR}/${TARBALL_NAME}" -C "${STAGE_DIR}" .
  rm -rf "${STAGE_DIR}"
done

rm -rf "${RELEASES_DIR}/stage"

echo "==> Generating SHA256 checksums..."
(
  cd "${RELEASES_DIR}"
  sha256sum straitgateway-v*.tar.gz > SHA256SUMS
)

echo "✓ Multi-arch builds and release packages created in ${RELEASES_DIR}/:"
ls -lh "${RELEASES_DIR}"/straitgateway-v*.tar.gz "${RELEASES_DIR}/SHA256SUMS"
