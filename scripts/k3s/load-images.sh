#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REGISTRY="ghcr.io/msaeedb40"
for img in straitgateway-controller straitgatewayd straitgateway-cli straitgateway-ui; do
  echo "[load] Importing ${REGISTRY}/${img}:latest into k3s..."
  ctr --namespace k8s.io images import <(docker save "${REGISTRY}/${img}:latest") 2>/dev/null || \
    k3s ctr images import <(docker save "${REGISTRY}/${img}:latest")
done
echo "[straitgateway] Images loaded into k3s."
