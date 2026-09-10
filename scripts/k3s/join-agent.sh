#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Run on worker nodes to join a k3s cluster.
set -euo pipefail
K3S_URL="${K3S_URL:?Set K3S_URL to https://<server>:6443}"
K3S_TOKEN="${K3S_TOKEN:?Set K3S_TOKEN from /var/lib/rancher/k3s/server/node-token}"
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="\
  --flannel-backend=none \
  --disable-network-policy \
  --disable-kube-proxy" \
  K3S_URL="${K3S_URL}" K3S_TOKEN="${K3S_TOKEN}" sh -
echo "[straitgateway] k3s agent joined cluster."
