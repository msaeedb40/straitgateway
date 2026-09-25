# StraitGateway Makefile
#
# Targets:
#   build         - Build all binaries
#   build-ebpf    - Compile eBPF programs with Clang 22 / LLVM 22
#   test          - Run unit tests
#   test-race     - Run unit tests with race detector
#   lint          - Run golangci-lint
#   fmt           - Format Go source files
#   vet           - Run go vet
#   generate      - Run code generation (deepcopy, CRD manifests)
#   kind-create   - Create kind cluster for development
#   kind-deploy   - Deploy StraitGateway to kind cluster
#   kind-delete   - Delete kind cluster
#   docker-build  - Build Docker images
#   helm-package  - Package the Helm chart
#   clean         - Remove build artifacts

SHELL := /bin/bash
.DEFAULT_GOAL := build

# ─── Version ───────────────────────────────────────────────────────────────────
VERSION       ?= 1.0.1
GIT_COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION    ?= $(shell go version | awk '{print $$3}')
PLATFORM      ?= $(shell go env GOOS)/$(shell go env GOARCH)

LDFLAGS := -ldflags "\
  -X github.com/straitgateway/straitgateway/pkg/version.Version=$(VERSION) \
  -X github.com/straitgateway/straitgateway/pkg/version.GitCommit=$(GIT_COMMIT) \
  -X github.com/straitgateway/straitgateway/pkg/version.BuildDate=$(BUILD_DATE) \
  -X github.com/straitgateway/straitgateway/pkg/version.GoVersion=$(GO_VERSION) \
  -X github.com/straitgateway/straitgateway/pkg/version.Platform=$(PLATFORM) \
  -w -s"

# ─── Directories ───────────────────────────────────────────────────────────────
DIST_DIR      := dist
AMD64_DIR     := $(DIST_DIR)/linux-amd64
ARM64_DIR     := $(DIST_DIR)/linux-arm64

# ─── eBPF toolchain ────────────────────────────────────────────────────────────
CLANG         ?= $(shell which clang-22 clang-21 clang 2>/dev/null | head -n1)
LLC           ?= $(shell which llc-22 llc-21 llc 2>/dev/null | head -n1)
BPFTOOL       ?= bpftool
BPF_SRC       := bpf
BPF_OUT       := bpf/out

# ─── Docker ────────────────────────────────────────────────────────────────────
REGISTRY      ?= ghcr.io/straitgateway
IMAGE_TAG     ?= $(VERSION)

# ─── Binaries ──────────────────────────────────────────────────────────────────
BINARIES := straitd sg-controller tgwd strait-cni sgctl sgpktcap

# ═══════════════════════════════════════════════════════════════════════════════
# Build
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: build
build: build-amd64 build-arm64

.PHONY: build-amd64
build-amd64:
	@mkdir -p $(AMD64_DIR)
	@echo "→ Building linux/amd64 binaries"
	@for bin in $(BINARIES); do \
	  echo "  building $$bin"; \
	  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
	    -o $(AMD64_DIR)/$$bin ./cmd/$$bin/ || exit 1; \
	done
	@echo "✓ linux/amd64 build complete"

.PHONY: build-arm64
build-arm64:
	@mkdir -p $(ARM64_DIR)
	@echo "→ Building linux/arm64 binaries"
	@for bin in $(BINARIES); do \
	  echo "  building $$bin"; \
	  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) \
	    -o $(ARM64_DIR)/$$bin ./cmd/$$bin/ || exit 1; \
	done
	@echo "✓ linux/arm64 build complete"

.PHONY: build-local
build-local:
	@echo "→ Building local binaries"
	@for bin in $(BINARIES); do \
	  echo "  building $$bin"; \
	  CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/$$bin ./cmd/$$bin/ || exit 1; \
	done
	@echo "✓ local build complete"

# ═══════════════════════════════════════════════════════════════════════════════
# eBPF
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: build-ebpf
build-ebpf:
	@echo "→ Compiling eBPF programs"
	@$(MAKE) -C $(BPF_SRC) CLANG=$(CLANG) LLC=$(LLC)
	@echo "✓ eBPF compilation complete"

.PHONY: generate-vmlinux
generate-vmlinux:
	@echo "→ Generating vmlinux.h"
	$(BPFTOOL) btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h
	@echo "✓ vmlinux.h generated"

# ═══════════════════════════════════════════════════════════════════════════════
# Code Generation
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: generate
generate: generate-deepcopy generate-crds generate-proto
	@echo "✓ all code generation complete"

.PHONY: generate-deepcopy
generate-deepcopy:
	@echo "→ Running controller-gen (deepcopy)"
	@which controller-gen > /dev/null 2>&1 || \
	  go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	controller-gen \
	  object:headerFile="tools/codegen/boilerplate.go.txt" \
	  paths="./api/..." \
	  output:object:dir="api/v1alpha4"
	@echo "✓ deepcopy generated"

.PHONY: generate-crds
generate-crds:
	@echo "→ Running controller-gen (CRD manifests)"
	@which controller-gen > /dev/null 2>&1 || \
	  go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	controller-gen \
	  crd \
	  paths="./api/..." \
	  output:crd:dir="deploy/crds"
	@echo "✓ CRD manifests generated → deploy/crds/"

.PHONY: generate-proto
generate-proto:
	@echo "→ Running protoc (straitd gRPC API)"
	@which protoc > /dev/null 2>&1 || (echo "protoc not found. Install: apt-get install -y protobuf-compiler"; exit 1)
	@which protoc-gen-go > /dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@which protoc-gen-go-grpc > /dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc \
	  --go_out=internal/tgwd/api \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=internal/tgwd/api \
	  --go-grpc_opt=paths=source_relative \
	  -I internal/tgwd/api \
	  internal/tgwd/api/straitd_api.proto
	@echo "✓ proto generated"

.PHONY: install-codegen-tools
install-codegen-tools:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✓ codegen tools installed"


# ═══════════════════════════════════════════════════════════════════════════════
# Test
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: test
test:
	@echo "→ Running unit tests"
	go test ./... -v -count=1 -timeout=120s
	@echo "✓ tests complete"

.PHONY: test-race
test-race:
	@echo "→ Running tests with race detector"
	go test ./... -race -count=1 -timeout=120s

.PHONY: test-integration
test-integration:
	@echo "→ Running integration tests (requires live straitd or kind cluster)"
	SG_INTEGRATION_TEST=1 go test ./test/integration/... -v -count=1 -timeout=300s

.PHONY: test-e2e
test-e2e:
	@echo "→ Running e2e tests (requires kind cluster with StraitGateway)"
	SG_E2E_TEST=1 go test ./test/e2e/... -v -count=1 -timeout=300s

.PHONY: test-ipam
test-ipam:
	@echo "→ Running IPAM unit tests"
	go test ./internal/cni/ipam/... -v -count=1 -timeout=30s
.PHONY: test-conformance
test-conformance:
	@echo "→ Running Gateway API conformance tests"
	./test/e2e/run-conformance.sh

# ═══════════════════════════════════════════════════════════════════════════════
# Quality
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: lint
lint:
	@echo "→ Running golangci-lint"
	golangci-lint run ./...

.PHONY: fmt
fmt:
	@echo "→ Formatting Go source"
	gofmt -w -s .
	goimports -w .

.PHONY: vet
vet:
	@echo "→ Running go vet"
	go vet ./...

.PHONY: tidy
tidy:
	@echo "→ Running go mod tidy"
	go mod tidy

# ═══════════════════════════════════════════════════════════════════════════════
# Kind (local dev)
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: kind-create
kind-create:
	@echo "→ Creating kind cluster"
	./scripts/kind/create-cluster.sh

.PHONY: kind-build
kind-build: build-amd64 build-ebpf
	@echo "→ Building and loading images into kind"
	./scripts/kind/build-kind.sh

.PHONY: kind-deploy
kind-deploy:
	@echo "→ Deploying StraitGateway to kind"
	./scripts/kind/deploy-kind.sh

.PHONY: kind-delete
kind-delete:
	@echo "→ Deleting kind cluster"
	./scripts/kind/delete-cluster.sh

.PHONY: kind-uninstall
kind-uninstall:
	@echo "→ Uninstalling StraitGateway from kind"
	./scripts/kind/uninstall-kind.sh

# ═══════════════════════════════════════════════════════════════════════════════
# Docker
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: docker-build
docker-build:
	@echo "→ Building Docker images"
	@for bin in straitd sg-controller tgwd; do \
	  echo "  building $(REGISTRY)/$$bin:$(IMAGE_TAG)"; \
	  docker build --platform linux/amd64,linux/arm64 \
	    --build-arg BINARY=$$bin \
	    -t $(REGISTRY)/$$bin:$(IMAGE_TAG) . || exit 1; \
	done
	@echo "✓ Docker images built"

.PHONY: docker-push
docker-push:
	@for bin in straitd sg-controller tgwd; do \
	  docker push $(REGISTRY)/$$bin:$(IMAGE_TAG); \
	done

# ═══════════════════════════════════════════════════════════════════════════════
# Helm
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: helm-package
helm-package:
	@echo "→ Packaging Helm chart"
	helm package straitgateway-helm/ -d dist/
	@echo "✓ Helm chart packaged"

.PHONY: helm-lint
helm-lint:
	helm lint straitgateway-helm/

# ═══════════════════════════════════════════════════════════════════════════════
# Documentation (GitHub Pages)
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: docs-build
docs-build:
	@echo "→ Building documentation site"
	mkdocs build --clean
	@echo "✓ Documentation built in site/"

.PHONY: docs-serve
docs-serve:
	@echo "→ Serving documentation locally on http://127.0.0.1:8000"
	mkdocs serve

# ═══════════════════════════════════════════════════════════════════════════════
# Clean
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: clean
clean:
	@echo "→ Cleaning build artifacts"
	rm -rf $(DIST_DIR)
	rm -rf $(BPF_OUT)
	@echo "✓ clean complete"

# ═══════════════════════════════════════════════════════════════════════════════
# Help
# ═══════════════════════════════════════════════════════════════════════════════

.PHONY: help
help:
	@echo "StraitGateway Makefile targets:"
	@echo ""
	@echo "  build           Build all binaries (amd64 + arm64)"
	@echo "  build-local     Build binaries for current platform"
	@echo "  build-ebpf      Compile eBPF programs (Clang 22)"
	@echo "  generate        Run code generation"
	@echo "  test            Run unit tests"
	@echo "  test-race       Run tests with race detector"
	@echo "  test-integration Run integration tests"
	@echo "  test-e2e        Run e2e tests on kind"
	@echo "  lint            Run golangci-lint"
	@echo "  fmt             Format Go source"
	@echo "  vet             Run go vet"
	@echo "  tidy            Run go mod tidy"
	@echo "  kind-create     Create kind cluster"
	@echo "  kind-build      Build + load images into kind"
	@echo "  kind-deploy     Deploy StraitGateway to kind"
	@echo "  kind-delete     Delete kind cluster"
	@echo "  docker-build    Build Docker images"
	@echo "  helm-package    Package Helm chart"
	@echo "  clean           Remove build artifacts"
