#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
echo "→ Running dataplane validation..."
go test -v -count=1 ./dataplane/... ./ebpf/...
echo "✓ Dataplane tests passed"
