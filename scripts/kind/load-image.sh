#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway}"
REGISTRY="${REGISTRY:-ghcr.io/msaeedb40}"
TAG="${TAG:-latest}"
for img in straitgateway-controller straitgatewayd straitgateway-cli straitgateway-ui; do
  echo "→ Loading ${REGISTRY}/${img}:${TAG} into kind cluster '${CLUSTER}'..."
  kind load docker-image "${REGISTRY}/${img}:${TAG}" --name "${CLUSTER}"
done
echo "✓ Images loaded into kind cluster '${CLUSTER}'"
