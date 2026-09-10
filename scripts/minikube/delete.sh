#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${MINIKUBE_PROFILE:-straitgateway}"
minikube delete --profile="${CLUSTER}"
echo "[straitgateway] Minikube cluster '${CLUSTER}' deleted."
