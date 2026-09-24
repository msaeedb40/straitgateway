#!/usr/bin/env bash
# scripts/kind/uninstall-kind.sh
# Uninstalls StraitGateway from the kind cluster.

set -euo pipefail
CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
NAMESPACE="${NAMESPACE:-kube-system}"       

echo "→ Setting kubectl context to kind-${CLUSTER_NAME}"
kubectl config use-context "kind-${CLUSTER_NAME}"

echo "→ Uninstalling StraitGateway Helm release"
helm uninstall straitgateway --namespace "${NAMESPACE}" || true

echo "→ Removing CRDs"
kubectl delete -f deploy/crds/ --ignore-not-found

echo "✓ StraitGateway uninstalled from kind cluster '${CLUSTER_NAME}'"
