#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
CLUSTER="${KIND_CLUSTER:-straitgateway-dev}"
kind delete cluster --name "$CLUSTER"
