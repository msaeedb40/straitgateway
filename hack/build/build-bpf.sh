#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Compile all eBPF C programs using bpf2go (cilium/ebpf).
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${REPO_ROOT}"
export BPF_CLANG="${BPF_CLANG:-clang-22}"
export BPF_CFLAGS="${BPF_CFLAGS:--O2 -g -Wall -Werror}"
echo "→ Generating eBPF Go bindings via bpf2go..."
go generate ./bpf/...
echo "✓ eBPF programs compiled"
