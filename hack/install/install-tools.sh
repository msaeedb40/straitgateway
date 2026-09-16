#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
echo "→ Installing development tools..."
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
echo "✓ Tools installed in \${GOPATH}/bin"
