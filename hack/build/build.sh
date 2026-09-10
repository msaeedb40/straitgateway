#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Build all straitgateway binaries for the current platform.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
OUTPUT="${OUTPUT:-dist}"
mkdir -p "${OUTPUT}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-s -w \
  -X github.com/msaeedb40/straitgateway/internal/version.Version=${VERSION} \
  -X github.com/msaeedb40/straitgateway/internal/version.Commit=${COMMIT} \
  -X github.com/msaeedb40/straitgateway/internal/version.BuildDate=${BUILD_DATE}"

for cmd in straitgatewayd sg-controller sg-cli; do
  echo "→ Building ${cmd}..."
  CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o "${OUTPUT}/${cmd}" "./cmd/${cmd}/"
done
echo "✓ Binaries written to ${OUTPUT}/"
