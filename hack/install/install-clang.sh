#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Installs clang-22 and llvm-22 on Ubuntu/Debian.
set -euo pipefail
CLANG_VER=22
wget -O - https://apt.llvm.org/llvm.sh | sudo bash -s -- ${CLANG_VER}
sudo apt-get install -y clang-${CLANG_VER} llvm-${CLANG_VER} libbpf-dev linux-headers-$(uname -r)
echo "clang-${CLANG_VER} installed: $(clang-${CLANG_VER} --version)"
