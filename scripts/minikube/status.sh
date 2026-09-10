#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
echo "=== Minikube Status ==="
minikube status --profile="${MINIKUBE_PROFILE:-straitgateway}"
echo ""
echo "=== straitgateway Pods ==="
kubectl get pods -n straitgateway-system -o wide
echo ""
echo "=== GatewayClass ==="
kubectl get gatewayclass 2>/dev/null || echo "No GatewayClass CRD"
