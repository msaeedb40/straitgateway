#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
export KUBECONFIG="${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system --create-namespace \
  --set global.clusterName=k3s \
  --set kubeProxyReplacement.enabled=true \
  --wait --timeout=5m
echo "[straitgateway] Installed on k3s."
