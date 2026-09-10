#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway}"
kubectl config use-context "kind-${CLUSTER}"
helm uninstall straitgateway --namespace straitgateway-system || true
kubectl delete namespace straitgateway-system --ignore-not-found
echo "✓ straitgateway uninstalled from kind cluster '${CLUSTER}'"
