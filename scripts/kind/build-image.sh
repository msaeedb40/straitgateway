#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REGISTRY="${REGISTRY:-ghcr.io/msaeedb40}"
TAG="${TAG:-latest}"
for img in straitgateway-controller straitgatewayd straitgateway-cli straitgateway-ui; do
  echo "→ Building ${REGISTRY}/${img}:${TAG}..."
  case "$img" in
    straitgateway-ui) docker build -t "${REGISTRY}/${img}:${TAG}" -f build/Dockerfile.ui ./ui ;;
    straitgateway-controller) docker build -t "${REGISTRY}/${img}:${TAG}" -f build/Dockerfile.sg-controller . ;;
    straitgatewayd) docker build -t "${REGISTRY}/${img}:${TAG}" -f build/Dockerfile.straitgatewayd . ;;
    straitgateway-cli) docker build -t "${REGISTRY}/${img}:${TAG}" -f build/Dockerfile.sg-cli . ;;
  esac
done
echo "✓ All images built"
