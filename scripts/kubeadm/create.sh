#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Creates a kubeadm multi-node cluster for straitgateway testing.
# Run on the control-plane node as root.
set -euo pipefail
POD_CIDR="${POD_CIDR:-10.244.0.0/16}"
SERVICE_CIDR="${SERVICE_CIDR:-10.96.0.0/12}"

kubeadm init \
  --pod-network-cidr="${POD_CIDR}" \
  --service-cidr="${SERVICE_CIDR}" \
  --skip-phases=addon/kube-proxy \
  "$@"

mkdir -p "$HOME/.kube"
cp /etc/kubernetes/admin.conf "$HOME/.kube/config"
echo "[straitgateway] Control plane initialized. Now install straitgateway: bash scripts/kubeadm/install.sh"
