#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Installs straitgateway into a kubeadm multi-node cluster.
# Prerequisites: kubeadm cluster with --pod-network-cidr and --skip-phases=addon/kube-proxy
set -euo pipefail
CLUSTER_NAME="${CLUSTER_NAME:-straitgateway}"

echo "[straitgateway] Installing Gateway API CRDs..."
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml

echo "[straitgateway] Installing Helm chart..."
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system --create-namespace \
  --set global.clusterName="${CLUSTER_NAME}" \
  --set kubeProxyReplacement.enabled=true \
  --set kubeProxyReplacement.mode=none \
  --set dataplane.overlay=Native \
  --wait --timeout=10m

echo "[straitgateway] Waiting for DaemonSet rollout..."
kubectl rollout status daemonset/straitgateway-agent -n straitgateway-system --timeout=5m

echo "[straitgateway] Installation complete."
kubectl get pods -n straitgateway-system
