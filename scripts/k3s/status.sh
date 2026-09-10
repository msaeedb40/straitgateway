#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
export KUBECONFIG="${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
echo "=== k3s Status ==="
k3s kubectl get nodes -o wide
echo ""
echo "=== straitgateway Pods ==="
kubectl get pods -n straitgateway-system -o wide
