#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${MINIKUBE_PROFILE:-straitgateway}"
kubectl config use-context "${CLUSTER}"
helm upgrade straitgateway ./straitgateway-helm \
  --namespace straitgateway-system --set global.clusterName="${CLUSTER}" --wait --timeout=5m
echo "✓ straitgateway upgraded in Minikube '${CLUSTER}'"
