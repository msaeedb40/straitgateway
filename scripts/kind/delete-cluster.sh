#!/usr/bin/env bash
# scripts/kind/delete-cluster.sh
# Deletes the StraitGateway kind development cluster.

set -euo pipefail
CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
echo "→ Deleting kind cluster: ${CLUSTER_NAME}"
kind delete cluster --name "${CLUSTER_NAME}"
echo "✓ kind cluster '${CLUSTER_NAME}' deleted"
