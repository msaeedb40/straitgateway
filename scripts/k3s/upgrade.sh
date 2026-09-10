#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
export KUBECONFIG="${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
helm upgrade straitgateway ./straitgateway-helm \
  --namespace straitgateway-system --set global.clusterName=k3s --wait --timeout=5m
echo "✓ straitgateway upgraded in k3s"
