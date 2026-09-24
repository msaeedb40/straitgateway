#!/usr/bin/env bash
# test/e2e/run-conformance.sh
# Runs Gateway API conformance tests against a live StraitGateway installation.
#
# Prerequisites:
#   - kind cluster with StraitGateway deployed
#   - KUBECONFIG pointing to the cluster
#   - go install sigs.k8s.io/gateway-api/conformance/...@latest

set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
NAMESPACE="${NAMESPACE:-default}"
GATEWAY_CLASS="${GATEWAY_CLASS:-straitgateway}"
REPORT_DIR="${REPORT_DIR:-test/e2e/reports}"

echo "=== Gateway API Conformance Tests ==="
echo "  Cluster:      ${CLUSTER_NAME}"
echo "  GatewayClass: ${GATEWAY_CLASS}"

kubectl config use-context "kind-${CLUSTER_NAME}"

mkdir -p "${REPORT_DIR}"

# Install Gateway API CRDs if not present
echo "→ Installing Gateway API CRDs"
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml || true

# Wait for GatewayClass to be accepted
echo "→ Waiting for GatewayClass '${GATEWAY_CLASS}' to be Accepted"
kubectl wait --for=condition=Accepted \
  gatewayclass/"${GATEWAY_CLASS}" \
  --timeout=60s || true

# Run Gateway API conformance suite
echo "→ Running conformance tests"
go test \
  sigs.k8s.io/gateway-api/conformance \
  -v \
  -run TestConformance \
  -gateway-class="${GATEWAY_CLASS}" \
  -supported-features="Gateway,HTTPRoute,GRPCRoute" \
  -report-output="${REPORT_DIR}/conformance-report.yaml" \
  -timeout=300s

echo "✓ Conformance tests complete. Report: ${REPORT_DIR}/conformance-report.yaml"
