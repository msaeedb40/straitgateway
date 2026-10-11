# ============================================================
# straitd — StraitGateway Data Plane Daemon
# Requires: privileged, BPF, NET_ADMIN, SYS_ADMIN
# ============================================================

# ── Stage 0: LLVM/eBPF toolchain base (cached independently) ──
FROM golang:1.27.1-bookworm AS bpf-toolchain

RUN apt-get update -q && apt-get install -y --no-install-recommends \
    clang \
    llvm \
    libbpf-dev \
    linux-headers-generic \
    && rm -rf /var/lib/apt/lists/*

# ── Stage 1: Go module cache (invalidated only on go.mod/go.sum change) ──
FROM bpf-toolchain AS go-deps

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ── Stage 2: eBPF compilation (invalidated only on bpf/** change) ──
FROM go-deps AS bpf-builder

COPY bpf/ ./bpf/
COPY Makefile ./

RUN make build-bpf

# ── Stage 3: Go binary build ──
FROM go-deps AS go-builder

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETARCH=amd64
ARG TARGETOS=linux

COPY --from=bpf-builder /build/bpf ./bpf
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags "-s -w \
      -X 'github.com/straitgateway/straitgateway/pkg/version.Version=${VERSION}' \
      -X 'github.com/straitgateway/straitgateway/pkg/version.GitCommit=${GIT_COMMIT}' \
      -X 'github.com/straitgateway/straitgateway/pkg/version.BuildDate=${BUILD_DATE}'" \
    -o /out/straitd ./cmd/straitd

# ── Stage 4: Runtime ──
# distroless/static has no shell — smallest possible attack surface
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="StraitGateway - straitd"
LABEL org.opencontainers.image.description="StraitGateway eBPF/XDP/TC data plane daemon"
LABEL org.opencontainers.image.source="https://github.com/straitgateway/straitgateway"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="StraitGateway"

# BPF objects — embedded in image, pinned maps live in the BPF FS volume
COPY --from=bpf-builder /build/bpf/tc/      /var/lib/straitgateway/bpf/tc/
COPY --from=bpf-builder /build/bpf/xdp/     /var/lib/straitgateway/bpf/xdp/
COPY --from=bpf-builder /build/bpf/netkit/  /var/lib/straitgateway/bpf/netkit/
COPY --from=bpf-builder /build/bpf/cgroup/  /var/lib/straitgateway/bpf/cgroup/
COPY --from=bpf-builder /build/bpf/service/ /var/lib/straitgateway/bpf/service/

COPY --from=go-builder /out/straitd /usr/local/bin/straitd

# ── Volume declarations ──
# /sys/fs/bpf          — BPF filesystem for pinned maps (hostPath, Bidirectional)
# /sys/kernel/debug    — BTF + tracing (hostPath, HostToContainer)
# /lib/modules         — Kernel modules (hostPath, read-only)
# /var/run/straitgateway — Unix sockets & PID file
VOLUME ["/sys/fs/bpf", "/sys/kernel/debug", "/lib/modules", "/var/run/straitgateway"]

# straitd MUST run as root — securityContext.privileged=true in DaemonSet
USER root

ENTRYPOINT ["/usr/local/bin/straitd"]
