#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Generate DeepCopy functions and CRD YAML using controller-gen.
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
CONTROLLER_GEN="${CONTROLLER_GEN:-$(go env GOPATH)/bin/controller-gen}"
if ! command -v "${CONTROLLER_GEN}" &>/dev/null; then
  echo "→ Installing controller-gen..."
  go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
fi
echo "→ Generating DeepCopy functions..."
"${CONTROLLER_GEN}" object:headerFile=hack/boilerplate/license_header.txt \
  paths="./api/..."

CRD_TEMPLATES_DIR="${REPO_ROOT}/straitgateway-helm/charts/crds/templates"
DIST_CRDS_DIR="${REPO_ROOT}/dist/crds"
mkdir -p "${CRD_TEMPLATES_DIR}" "${DIST_CRDS_DIR}"

echo "→ Generating CRD YAML..."
"${CONTROLLER_GEN}" crd:generateEmbeddedObjectMeta=true \
  paths="./api/..." \
  output:crd:dir="${CRD_TEMPLATES_DIR}"

# Normalize filenames from straitgateway.io_<plural>.yaml to <plural>.yaml
for f in "${CRD_TEMPLATES_DIR}"/straitgateway.io_*.yaml; do
  if [[ -f "$f" ]]; then
    mv "$f" "$(echo "$f" | sed 's/straitgateway.io_//')"
  fi
done

echo "→ Generating standalone CRD bundle in ${DIST_CRDS_DIR}/straitgateway-crds.yaml..."
cat << 'HEADER' > "${DIST_CRDS_DIR}/straitgateway-crds.yaml"
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Straitgateway CustomResourceDefinitions (All-in-One Standalone Bundle)
# Apply directly via: kubectl apply -f straitgateway-crds.yaml
HEADER

for crd in "${CRD_TEMPLATES_DIR}"/*.yaml; do
  echo "---" >> "${DIST_CRDS_DIR}/straitgateway-crds.yaml"
  cat "$crd" >> "${DIST_CRDS_DIR}/straitgateway-crds.yaml"
done

echo "✓ CRD generation and bundling complete"
