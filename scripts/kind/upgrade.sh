#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway}"
kubectl config use-context "kind-${CLUSTER}"
helm upgrade straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --set global.clusterName="${CLUSTER}" \
  --wait --timeout=5m
echo "✓ straitgateway upgraded in kind cluster '${CLUSTER}'"
