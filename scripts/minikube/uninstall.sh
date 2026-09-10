#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
helm uninstall straitgateway --namespace straitgateway-system || true
kubectl delete namespace straitgateway-system --ignore-not-found
echo "[straitgateway] Uninstalled from Minikube."
