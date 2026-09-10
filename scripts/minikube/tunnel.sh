#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
echo "[straitgateway] Starting Minikube tunnel for LoadBalancer services..."
minikube tunnel --profile="${MINIKUBE_PROFILE:-straitgateway}"
