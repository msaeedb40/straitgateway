// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package tools

import (
	_ "github.com/cilium/ebpf/cmd/bpf2go"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
