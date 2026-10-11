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
all: build-bpf generate build

.PHONY: build-bpf
build-bpf:
	@echo "Compiling eBPF programs with Clang..."
	$(CLANG) $(CLANG_FLAGS) ./bpf/tc/ingress.c -o ./bpf/tc/ingress.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/tc/egress.c -o ./bpf/tc/egress.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/xdp/ingress.c -o ./bpf/xdp/ingress.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/xdp/redirect.c -o ./bpf/xdp/redirect.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/netkit/peer_ingress.c -o ./bpf/netkit/peer_ingress.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/netkit/peer_egress.c -o ./bpf/netkit/peer_egress.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/cgroup/connect4.c -o ./bpf/cgroup/connect4.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/cgroup/connect6.c -o ./bpf/cgroup/connect6.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/cgroup/sendmsg4.c -o ./bpf/cgroup/sendmsg4.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/cgroup/sendmsg6.c -o ./bpf/cgroup/sendmsg6.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/service/lb.c -o ./bpf/service/lb.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/service/nat.c -o ./bpf/service/nat.o
	$(CLANG) $(CLANG_FLAGS) ./bpf/service/conntrack.c -o ./bpf/service/conntrack.o

# Phase 13: RoCEv2 eBPF programs
.PHONY: build-bpf-roce
build-bpf-roce:
	@echo "Compiling RoCEv2 eBPF programs..."
	$(CLANG) $(CLANG_FLAGS) -I./bpf -I./bpf/include ./bpf/roce/roce_xdp.c -o ./bpf/roce/roce_xdp.o
	$(CLANG) $(CLANG_FLAGS) -I./bpf -I./bpf/include ./bpf/roce/roce_tc.c  -o ./bpf/roce/roce_tc.o

.PHONY: build
build: straitd sg-controller sg-cni sg-agent sgctl

.PHONY: build-ui
build-ui:
	@echo "Building UI Plane..."
	@cd ui && npm run build

.PHONY: build-all
build-all: build-bpf build build-ui

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

# ---------------------------------------------------------------------------
# Phase 12: Production Chaos Engineering & Conformance Automation
# ---------------------------------------------------------------------------

FUZZTIME ?= 30s
GATEWAY_CLASS ?= straitgateway
REPORT_DIR ?= _reports

.PHONY: conformance
conformance:
	@echo "Running Gateway API + CNI conformance tests..."
	@mkdir -p $(REPORT_DIR)
	$(GO) test -v -count=1 -tags conformance ./test/conformance/... \
	  -gateway-class $(GATEWAY_CLASS) \
	  -report-output $(REPORT_DIR)/gateway_conformance.json

.PHONY: conformance-cni
conformance-cni:
	@echo "Running CNI-only conformance tests (requires root)..."
	@mkdir -p $(REPORT_DIR)
	$(GO) test -v -count=1 -run "TestCNI" ./test/conformance/...

.PHONY: sonobuoy-run
sonobuoy-run:
	@echo "Submitting Sonobuoy conformance job to cluster..."
	sonobuoy run \
	  --plugin test/conformance/sonobuoy-plugin.yaml \
	  --wait \
	  --kubeconfig $(KUBECONFIG)

.PHONY: sonobuoy-retrieve
sonobuoy-retrieve:
	@mkdir -p $(REPORT_DIR)
	sonobuoy retrieve $(REPORT_DIR)/
	sonobuoy results $(REPORT_DIR)/*.tar.gz

.PHONY: chaos
chaos:
	@echo "Running chaos smoke tests (requires root for BPF/netem)..."
	bash hack/run-chaos.sh

.PHONY: chaos-enomem
chaos-enomem:
	@echo "Injecting ENOMEM fault for 2s..."
	$(GO) test -v -count=1 -run "TestInjectENOMEM" ./test/chaos/...

.PHONY: chaos-drop
chaos-drop:
	@echo "Injecting drop burst (32 pkts, 100%% probability)..."
	$(GO) test -v -count=1 -run "TestInjectDropBurst" ./test/chaos/...

.PHONY: fuzz
fuzz:
	@echo "Running verifier fuzzing matrix (fuzztime=$(FUZZTIME))..."
	FUZZTIME=$(FUZZTIME) bash hack/kernel-matrix-fuzz.sh

.PHONY: fuzz-unit
fuzz-unit:
	@echo "Running verifier fuzzer unit tests (no BPF required)..."
	$(GO) test -v -count=1 \
	  -run "TestKernelVersionParsing|TestKernelMatrixCompleteness|TestBuiltinSeeds|TestMutateInstructions|TestCorpusAddAndRetrieve" \
	  ./test/fuzz/...

.PHONY: test-phase12
test-phase12: fuzz-unit chaos
	@echo "Phase 12 non-privileged tests complete."

# ---------------------------------------------------------------------------
# Phase 13: AI/ML Workloads & RDMA/RoCEv2
# ---------------------------------------------------------------------------

.PHONY: test-roce
test-roce:
	@echo "Running RoCEv2 unit tests (no hardware required)..."
	$(GO) test -v -count=1 -race ./internal/dataplane/roce/...

.PHONY: test-roce-dcqcn
test-roce-dcqcn:
	@echo "Running DCQCN rate controller tests..."
	$(GO) test -v -count=1 -run "TestDCQCN" ./internal/dataplane/roce/...

.PHONY: build-roce
build-roce: build-bpf-roce
	@echo "RoCEv2 BPF objects compiled."

.PHONY: test-phase13
test-phase13: test-roce
	@echo "Phase 13 tests complete."

# ---------------------------------------------------------------------------
# Phase 14: Autonomous Dataplane AIOps
# ---------------------------------------------------------------------------

.PHONY: test-aiops
test-aiops:
	@echo "Running AIOps unit tests (burst detector, engine, trace context)..."
	$(GO) test -v -count=1 -race -timeout 90s ./internal/observability/aiops/...

.PHONY: test-aiops-burst
test-aiops-burst:
	@echo "Running BurstDetector tests only..."
	$(GO) test -v -count=1 -race -timeout 90s \
	  -run "TestBurstDetector" ./internal/observability/aiops/...

.PHONY: test-aiops-trace
test-aiops-trace:
	@echo "Running W3C Trace Context tests only..."
	$(GO) test -v -count=1 \
	  -run "TestTraceparent|TestSpanContext|TestPropagator|TestFlowKey" \
	  ./internal/observability/aiops/...

.PHONY: test-phase14
test-phase14: test-aiops
	@echo "Phase 14 tests complete."
