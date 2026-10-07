# StraitGateway Makefile
# Project: StraitGateway (Kubernetes-native Transit Gateway, CNI, & eBPF Dataplane)

SHELL := /bin/bash
GO ?= go
CLANG ?= clang
LLVM_STRIP ?= llvm-strip
BPF2GO ?= $(shell which bpf2go 2>/dev/null || echo $(GOPATH)/bin/bpf2go)
GOPATH ?= $(shell $(GO) env GOPATH)

VERSION ?= 1.0.0
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS ?= -X 'github.com/straitgateway/straitgateway/pkg/version.Version=$(VERSION)' \
           -X 'github.com/straitgateway/straitgateway/pkg/version.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/straitgateway/straitgateway/pkg/version.BuildDate=$(BUILD_DATE)'

BIN_DIR ?= bin
BPF_DIR ?= bpf
BPF_INCLUDE ?= $(BPF_DIR)/include

CLANG_FLAGS := -O2 -g -Wall -Werror \
               -target bpf \
               -D__TARGET_ARCH_x86 \
               -I$(BPF_INCLUDE) \
               -c

.PHONY: all
all: generate build

.PHONY: build
build: straitd sg-controller sg-cni sg-agent sgctl

.PHONY: straitd
straitd:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/straitd ./cmd/straitd

.PHONY: sg-controller
sg-controller:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/sg-controller ./cmd/sg-controller

.PHONY: sg-cni
sg-cni:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/sg-cni ./cmd/sg-cni

.PHONY: sg-agent
sg-agent:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/sg-agent ./cmd/sg-agent

.PHONY: sgctl
sgctl:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/sgctl ./cmd/sgctl

.PHONY: generate
generate:
	@echo "Running go generate..."
	$(GO) generate ./...

.PHONY: test
test:
	$(GO) test -v -race ./pkg/... ./internal/...

.PHONY: lint
lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || $(GO) vet ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
