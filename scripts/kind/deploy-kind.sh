#!/usr/bin/env bash
# scripts/kind/deploy-kind.sh
# Deploys StraitGateway to the kind cluster via Helm.

set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
VERSION="${VERSION:-dev}"
NAMESPACE="${NAMESPACE:-kube-system}"

echo "→ Setting kubectl context to kind-${CLUSTER_NAME}"
kubectl config use-context "kind-${CLUSTER_NAME}"

echo "→ Applying CRDs"
kubectl apply -f deploy/crds/

echo "→ Deploying StraitGateway via Helm"
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --set global.tag="${VERSION}" \
  --set global.registry="straitgateway" \
  --set global.pullPolicy=Never \
  --wait \
  --timeout=120s

echo "✓ StraitGateway deployed to kind cluster '${CLUSTER_NAME}'"
echo "  Run: kubectl -n ${NAMESPACE} get pods"
