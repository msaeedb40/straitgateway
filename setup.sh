#!/usr/bin/env bash
# setup.sh — StraitGateway developer environment bootstrap
# Sets up toolchain requirements for local development.

set -euo pipefail

echo "=== StraitGateway Developer Setup ==="

# ─── Go ────────────────────────────────────────────────────────────────────────
GO_VERSION="1.27.1"
if ! command -v go &>/dev/null || [[ "$(go version | awk '{print $3}')" != "go${GO_VERSION}" ]]; then
  echo "→ Installing Go ${GO_VERSION}"
  wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  echo "✓ Go ${GO_VERSION} installed"
else
  echo "✓ Go $(go version | awk '{print $3}') already installed"
fi

# ─── kind ──────────────────────────────────────────────────────────────────────
if ! command -v kind &>/dev/null; then
  echo "→ Installing kind"
  go install sigs.k8s.io/kind@latest
  echo "✓ kind installed"
else
  echo "✓ kind $(kind --version) already installed"
fi

# ─── kubectl ───────────────────────────────────────────────────────────────────
if ! command -v kubectl &>/dev/null; then
  echo "→ Installing kubectl"
  curl -sLO "https://dl.k8s.io/release/$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl && sudo mv kubectl /usr/local/bin/
  echo "✓ kubectl installed"
else
  echo "✓ kubectl $(kubectl version --client --short 2>/dev/null | head -1) already installed"
fi

# ─── Helm ──────────────────────────────────────────────────────────────────────
if ! command -v helm &>/dev/null; then
  echo "→ Installing Helm"
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  echo "✓ Helm installed"
else
  echo "✓ Helm $(helm version --short) already installed"
fi

# ─── golangci-lint ─────────────────────────────────────────────────────────────
if ! command -v golangci-lint &>/dev/null; then
  echo "→ Installing golangci-lint"
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  echo "✓ golangci-lint installed"
else
  echo "✓ golangci-lint already installed"
fi

# ─── bpftool ───────────────────────────────────────────────────────────────────
if ! command -v bpftool &>/dev/null; then
  echo "→ bpftool not found; install manually or via package manager"
  echo "  sudo apt-get install -y bpftool"
fi

# ─── Clang 22 / LLVM 22 ────────────────────────────────────────────────────────
if ! command -v clang-22 &>/dev/null; then
  echo "→ Clang 22 not found"
  echo "  To install: wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | sudo apt-key add -"
  echo "  echo 'deb http://apt.llvm.org/noble/ llvm-toolchain-noble-22 main' | sudo tee /etc/apt/sources.list.d/llvm.list"
  echo "  sudo apt-get update && sudo apt-get install -y clang-22 llvm-22"
else
  echo "✓ clang-22 already installed"
fi

echo ""
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  make kind-create    # Create dev cluster"
echo "  make kind-build     # Build + load images"
echo "  make kind-deploy    # Deploy StraitGateway"
echo "  sgctl status        # Check status"
