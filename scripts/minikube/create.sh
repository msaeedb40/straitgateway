#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${MINIKUBE_PROFILE:-straitgateway}"
minikube start \
  --profile="${CLUSTER}" \
  --network-plugin=cni \
  --cni=false \
  --extra-config=kubeadm.skip-phases=addon/kube-proxy \
  --container-runtime=containerd \
  --memory=4096 \
  --cpus=2 \
  --nodes=2
echo "[straitgateway] Minikube cluster '${CLUSTER}' created — CNI disabled, kube-proxy skipped."
