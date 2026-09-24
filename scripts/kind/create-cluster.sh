#!/usr/bin/env bash
# scripts/kind/create-cluster.sh
# Creates a kind cluster for StraitGateway development.
#
# The cluster is configured with:
#   - disableDefaultCNI: true  (StraitGateway provides the CNI)
#   - kubeProxyMode: "none"    (StraitGateway replaces kube-proxy)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
KIND_CONFIG="${SCRIPT_DIR}/kind-config.yaml"

echo "→ Creating kind cluster: ${CLUSTER_NAME}"
kind create cluster \
  --name "${CLUSTER_NAME}" \
  --config "${KIND_CONFIG}" \
  --wait 300s

echo "✓ kind cluster '${CLUSTER_NAME}' created"
echo "  kubectl context: kind-${CLUSTER_NAME}"
