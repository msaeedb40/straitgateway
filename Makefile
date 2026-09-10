# Copyright 2026 straitgateway Authors
# SPDX-License-Identifier: Apache-2.0

# ============================================================
# Variables
# ============================================================
MODULE      := github.com/msaeedb40/straitgateway
REGISTRY    := ghcr.io/msaeedb40
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOARCH      ?= $(shell go env GOARCH)
GOOS        ?= $(shell go env GOOS)

LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) \
            -X $(MODULE)/internal/version.Commit=$(COMMIT) \
            -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

BPF_CLANG   ?= clang-22
BPF_STRIP   ?= llvm-strip-22
BPFTOOL     ?= bpftool
LIBBPF_INCLUDE ?= /usr/include/bpf

PLATFORMS   := linux/amd64 linux/arm64

HELM_CHART  := straitgateway-helm
HELM_DIST   := dist/charts

KIND_CLUSTER   := straitgateway-dev
MINIKUBE_PROFILE := straitgateway-dev

IMAGE_CONTROLLER := $(REGISTRY)/straitgateway-controller:$(VERSION)
IMAGE_DAEMON     := $(REGISTRY)/straitgatewayd:$(VERSION)
IMAGE_CLI        := $(REGISTRY)/straitgateway-cli:$(VERSION)
IMAGE_UI         := $(REGISTRY)/straitgateway-ui:$(VERSION)

# ============================================================
# Default
# ============================================================
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ============================================================
# Generate
# ============================================================
.PHONY: generate
generate: generate-deepcopy generate-crds generate-bpf ## Run all code generators

.PHONY: generate-deepcopy
generate-deepcopy: ## Generate DeepCopy methods
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
	  object:headerFile=hack/boilerplate/license_header.txt \
	  paths="./api/..."

.PHONY: generate-crds
generate-crds: ## Generate CRD YAML manifests
	go run sigs.k8s.io/controller-tools/cmd/controller-gen \
	  crd:maxDescLen=0 \
	  paths="./api/..." \
	  output:crd:dir=$(HELM_CHART)/charts/crds

.PHONY: generate-bpf
generate-bpf: ## Generate eBPF Go bindings with bpf2go
	BPF_CLANG=$(BPF_CLANG) go generate ./ebpf/...

# ============================================================
# Build
# ============================================================
.PHONY: build
build: build-controller build-daemon build-cli ## Build all Go binaries

.PHONY: build-controller
build-controller: ## Build sg-controller binary
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	  -ldflags "$(LDFLAGS)" \
	  -o bin/sg-controller \
	  ./cmd/sg-controller

.PHONY: build-daemon
build-daemon: ## Build straitgatewayd binary
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	  -ldflags "$(LDFLAGS)" \
	  -o bin/straitgatewayd \
	  ./cmd/straitgatewayd

.PHONY: build-cli
build-cli: ## Build sg-cli binary
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	  -ldflags "$(LDFLAGS)" \
	  -o bin/sg-cli \
	  ./cmd/sg-cli

.PHONY: build-bpf
build-bpf: ## Compile eBPF C programs
	$(MAKE) -C bpf all

.PHONY: build-all-arch
build-all-arch: ## Build binaries for all architectures
	@for platform in $(PLATFORMS); do \
	  os=$$(echo $$platform | cut -d/ -f1); \
	  arch=$$(echo $$platform | cut -d/ -f2); \
	  echo "Building $$os/$$arch..."; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build \
	    -ldflags "$(LDFLAGS)" \
	    -o bin/sg-controller-$$arch \
	    ./cmd/sg-controller; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build \
	    -ldflags "$(LDFLAGS)" \
	    -o bin/straitgatewayd-$$arch \
	    ./cmd/straitgatewayd; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build \
	    -ldflags "$(LDFLAGS)" \
	    -o bin/sg-cli-$$arch \
	    ./cmd/sg-cli; \
	done

# ============================================================
# Docker Images
# ============================================================
.PHONY: docker-build
docker-build: ## Build all Docker images (current arch)
	docker build -f build/Dockerfile.sg-controller \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_CONTROLLER) .
	docker build -f build/Dockerfile.straitgatewayd \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_DAEMON) .
	docker build -f build/Dockerfile.sg-cli \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_CLI) .
	docker build -f build/Dockerfile.ui \
	  -t $(IMAGE_UI) .

.PHONY: docker-push
docker-push: docker-build ## Push images to registry
	docker push $(IMAGE_CONTROLLER)
	docker push $(IMAGE_DAEMON)
	docker push $(IMAGE_CLI)
	docker push $(IMAGE_UI)

.PHONY: docker-buildx
docker-buildx: ## Build and push multi-arch images via buildx
	docker buildx build --platform linux/amd64,linux/arm64 \
	  -f build/Dockerfile.sg-controller \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_CONTROLLER) --push .
	docker buildx build --platform linux/amd64,linux/arm64 \
	  -f build/Dockerfile.straitgatewayd \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_DAEMON) --push .
	docker buildx build --platform linux/amd64,linux/arm64 \
	  -f build/Dockerfile.sg-cli \
	  --build-arg VERSION=$(VERSION) \
	  -t $(IMAGE_CLI) --push .
	docker buildx build --platform linux/amd64,linux/arm64 \
	  -f build/Dockerfile.ui \
	  -t $(IMAGE_UI) --push .

# ============================================================
# Test
# ============================================================
.PHONY: test
test: test-unit ## Run all tests

.PHONY: test-unit
test-unit: ## Run unit tests with race detector
	go test -v -race -count=1 ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	bash hack/test/test-integration.sh

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	bash hack/test/test-e2e.sh

.PHONY: test-dataplane
test-dataplane: ## Run dataplane/BPF verifier tests
	bash hack/test/test-dataplane.sh

# ============================================================
# Lint / Verify
# ============================================================
.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format Go source
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: verify
verify: ## Run all verification checks
	bash hack/verify/verify-gofmt.sh
	bash hack/verify/verify-govet.sh
	bash hack/verify/verify-generated.sh
	bash hack/verify/verify-crds.sh
	bash hack/verify/verify-bpf.sh
	bash hack/verify/verify-manifests.sh

# ============================================================
# Helm
# ============================================================
.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint $(HELM_CHART)/

.PHONY: helm-template
helm-template: ## Render Helm templates (dry-run)
	helm template straitgateway $(HELM_CHART)/ --dry-run

.PHONY: helm-package
helm-package: ## Package the Helm chart
	mkdir -p $(HELM_DIST)
	helm package $(HELM_CHART)/ -d $(HELM_DIST)

.PHONY: helm-index
helm-index: helm-package ## Build Helm repo index
	helm repo index $(HELM_DIST) --url https://charts.straitgateway.io

.PHONY: helm-serve
helm-serve: helm-index ## Serve Helm repo locally for testing
	cd $(HELM_DIST) && python3 -m http.server 8080

# ============================================================
# kind
# ============================================================
.PHONY: kind-create
kind-create: ## Create kind development cluster
	bash scripts/kind/create-cluster.sh

.PHONY: kind-delete
kind-delete: ## Delete kind development cluster
	bash scripts/kind/delete-cluster.sh

.PHONY: kind-load
kind-load: docker-build ## Load images into kind cluster
	bash scripts/kind/load-image.sh $(IMAGE_CONTROLLER) $(IMAGE_DAEMON) $(IMAGE_CLI)

.PHONY: kind-install
kind-install: kind-load ## Install straitgateway into kind cluster
	bash scripts/kind/install.sh

.PHONY: kind-upgrade
kind-upgrade: kind-load ## Upgrade straitgateway in kind cluster
	bash scripts/kind/upgrade.sh

.PHONY: kind-uninstall
kind-uninstall: ## Uninstall straitgateway from kind cluster
	bash scripts/kind/uninstall.sh

# ============================================================
# minikube
# ============================================================
.PHONY: minikube-create
minikube-create: ## Create minikube development cluster
	bash scripts/minikube/create-cluster.sh

.PHONY: minikube-load
minikube-load: docker-build ## Load images into minikube
	bash scripts/minikube/load-image.sh $(IMAGE_CONTROLLER) $(IMAGE_DAEMON) $(IMAGE_CLI)

.PHONY: minikube-install
minikube-install: minikube-load ## Install straitgateway into minikube
	bash scripts/minikube/install.sh

# ============================================================
# k3s
# ============================================================
.PHONY: k3s-install
k3s-install: ## Install straitgateway into k3s cluster
	bash scripts/k3s/install.sh

# ============================================================
# kubeadm
# ============================================================
.PHONY: kubeadm-install
kubeadm-install: ## Install straitgateway into kubeadm cluster
	bash scripts/kubeadm/install.sh

# ============================================================
# Install tools
# ============================================================
.PHONY: install-tools
install-tools: ## Install development tools
	bash hack/install/install-tools.sh

.PHONY: install-clang
install-clang: ## Install clang-22 / llvm-22
	bash hack/install/install-clang.sh

# ============================================================
# Clean
# ============================================================
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ dist/ bpf/*.o ebpf/*_bpfel.go ebpf/*_bpfeb.go

.PHONY: clean-all
clean-all: clean ## Remove all generated files
	rm -rf straitgateway-helm/charts/crds/*.yaml
