#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Installs straitgateway into the kind cluster via Helm.
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway-dev}"
kubectl config use-context "kind-${CLUSTER}"
# Install Gateway API CRDs (standard channel)
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml
# Install straitgateway chart
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system --create-namespace \
  --set global.clusterName="${CLUSTER}" \
  --set kubeProxyReplacement.enabled=true \
  --wait --timeout=5m
echo "straitgateway installed in kind cluster '${CLUSTER}'"
