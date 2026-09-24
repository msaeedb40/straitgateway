#!/usr/bin/env bash
# tools/codegen/generate.sh
# Runs controller-gen to regenerate CRD manifests and deepcopy methods.
# Also runs protoc to regenerate gRPC code from the straitd_api.proto.
#
# Prerequisites:
#   go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#   apt-get install -y protobuf-compiler

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=== StraitGateway Code Generation ==="

# ─── controller-gen: deepcopy + CRD manifests ──────────────────────────────────
echo "→ Running controller-gen (deepcopy + CRDs)"
controller-gen \
  object:headerFile="${ROOT}/tools/codegen/boilerplate.go.txt" \
  paths="${ROOT}/api/..." \
  output:object:dir="${ROOT}/api/v1alpha4"

controller-gen \
  crd:trivialVersions=true \
  paths="${ROOT}/api/..." \
  output:crd:dir="${ROOT}/deploy/crds"

echo "✓ controller-gen complete"

# ─── protoc: straitd API ───────────────────────────────────────────────────────
PROTO_SRC="${ROOT}/internal/tgwd/api/straitd_api.proto"
PROTO_OUT="${ROOT}/internal/tgwd/api"

echo "→ Running protoc (straitd API)"
protoc \
  --go_out="${PROTO_OUT}" \
  --go_opt=paths=source_relative \
  --go-grpc_out="${PROTO_OUT}" \
  --go-grpc_opt=paths=source_relative \
  -I "${ROOT}/internal/tgwd/api" \
  "${PROTO_SRC}"

echo "✓ protoc complete"
echo "=== Code generation done ==="
