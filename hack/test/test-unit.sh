#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
set -euo pipefail
go test -v -race -count=1 -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
