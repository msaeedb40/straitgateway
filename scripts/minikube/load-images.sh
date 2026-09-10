#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${MINIKUBE_PROFILE:-straitgateway}"
REGISTRY="ghcr.io/msaeedb40"
for img in straitgateway-controller straitgatewayd straitgateway-cli straitgateway-ui; do
  echo "[load] ${REGISTRY}/${img}:latest"
  minikube image load "${REGISTRY}/${img}:latest" --profile="${CLUSTER}" 2>/dev/null || \
    docker pull "${REGISTRY}/${img}:latest" && minikube image load "${REGISTRY}/${img}:latest" --profile="${CLUSTER}"
done
echo "[straitgateway] Images loaded into Minikube."
