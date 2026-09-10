#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Creates a k3s cluster configured for straitgateway.
set -euo pipefail
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="\
  --flannel-backend=none \
  --disable-network-policy \
  --disable=traefik \
  --disable=servicelb \
  --disable-kube-proxy \
  --write-kubeconfig-mode=644" sh -
echo "[straitgateway] k3s installed — flannel/kube-proxy/traefik disabled."
echo "Export KUBECONFIG: export KUBECONFIG=/etc/rancher/k3s/k3s.yaml"
