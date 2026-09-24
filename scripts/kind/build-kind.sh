#!/usr/bin/env bash
# scripts/kind/build-kind.sh
# Build StraitGateway binaries, build Docker images,
# and load them into the kind cluster.
# Also mounts eBPF programs and validates Netkit support.

set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-straitgateway-cluster}"
VERSION="${VERSION:-dev}"
REGISTRY="straitgateway"

echo "→ Building linux/amd64 binaries"
make build-amd64 VERSION="${VERSION}" 

echo "→ Building Docker images"
for bin in straitd sg-controller tgwd; do
  echo "  building ${REGISTRY}/${bin}:${VERSION}"
  docker build \
    --platform linux/amd64 \
    --build-arg BINARY="${bin}" \
    -t "${REGISTRY}/${bin}:${VERSION}" .
done

echo "→ Loading images into kind cluster: ${CLUSTER_NAME}"
for bin in straitd sg-controller tgwd; do
  kind load docker-image "${REGISTRY}/${bin}:${VERSION}" \
    --name "${CLUSTER_NAME}"
done

echo "→ Copying CNI binary into kind nodes"
for node in $(kind get nodes --name "${CLUSTER_NAME}"); do
  echo "  node: ${node}"
  docker cp dist/linux-amd64/strait-cni "${node}:/opt/cni/bin/strait-cni"
  docker exec "${node}" chmod +x /opt/cni/bin/strait-cni
done

echo "✓ build-kind.sh complete"
