#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Creates a kind cluster configured for straitgateway.
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway-dev}"
cat <<KINDCFG | kind create cluster --name "$CLUSTER" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true     # straitgateway manages CNI
  kubeProxyMode: none         # straitgateway replaces kube-proxy
nodes:
  - role: control-plane
  - role: worker
  - role: worker
KINDCFG
echo "kind cluster '${CLUSTER}' created — kube-proxy disabled, CNI disabled."
