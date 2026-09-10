#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
echo "[e2e] Creating kind cluster..."
bash scripts/kind/create-cluster.sh
echo "[e2e] Installing straitgateway..."
bash scripts/kind/install.sh
echo "[e2e] Running Go e2e tests..."
go test -v -timeout=30m ./test/e2e/... 2>&1 | tee /tmp/e2e-results.txt
echo "[e2e] Cleaning up..."
bash scripts/kind/delete-cluster.sh
