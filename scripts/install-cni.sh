#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
# Copies straitgateway CNI binary and config to host paths.
set -euo pipefail
CNI_BIN_DIR="${CNI_BIN_DIR:-/host/opt/cni/bin}"
CNI_CONF_DIR="${CNI_CONF_DIR:-/host/etc/cni/net.d}"
cp /opt/cni/bin/straitgateway "${CNI_BIN_DIR}/straitgateway"
cat > "${CNI_CONF_DIR}/10-straitgateway.conflist" << CNICONF
{
  "cniVersion": "1.1.0",
  "name": "straitgateway",
  "plugins": [
    {
      "type": "straitgateway",
      "socketPath": "/run/straitgateway/daemon.sock",
      "enableIPv4": true,
      "enableIPv6": false
    }
  ]
}
CNICONF
echo "[install-cni] straitgateway CNI installed to ${CNI_BIN_DIR} and ${CNI_CONF_DIR}"
