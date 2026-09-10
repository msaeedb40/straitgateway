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

echo "→ Generating CRD YAML..."
"${CONTROLLER_GEN}" crd:generateEmbeddedObjectMeta=true \
  paths="./api/..." \
  output:crd:artifacts:config=straitgateway-helm/charts/crds/

echo "✓ Code generation complete"
