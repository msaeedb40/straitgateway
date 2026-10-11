# ============================================================
# sg-agent — StraitGateway Node Agent (DaemonSet)
# Requires: privileged, BPF, NET_ADMIN, SYS_ADMIN, SYS_RESOURCE
# ============================================================

# ── Stage 0: LLVM/eBPF toolchain base ──
FROM golang:1.27.1-bookworm AS bpf-toolchain

RUN apt-get update -q && apt-get install -y --no-install-recommends \
    clang \
    llvm \
    libbpf-dev \
    linux-headers-generic \
    && rm -rf /var/lib/apt/lists/*

# ── Stage 1: Go module cache ──
FROM bpf-toolchain AS go-deps

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ── Stage 2: eBPF compilation ──
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
    -o /out/sg-agent ./cmd/sg-agent

# ── Stage 4: Runtime ──
# debian-slim (not distroless) — iproute2, wg, iptables needed for
# WireGuard tunnel setup, BFD link management, and BGP socket operations.
FROM debian:bookworm-slim AS runtime

LABEL org.opencontainers.image.title="StraitGateway - sg-agent"
LABEL org.opencontainers.image.description="StraitGateway node agent — eBPF loader, BGP, BFD, WireGuard"
LABEL org.opencontainers.image.source="https://github.com/straitgateway/straitgateway"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="StraitGateway"

RUN apt-get update -q && apt-get install -y --no-install-recommends \
    iproute2 \
    iptables \
    nftables \
    wireguard-tools \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    # Remove setuid/setgid bits from all apt-installed binaries
    && find /usr /bin /sbin -xdev -perm /6000 -exec chmod a-s {} + 2>/dev/null || true

# BPF objects
COPY --from=bpf-builder /build/bpf/tc/      /var/lib/straitgateway/bpf/tc/
COPY --from=bpf-builder /build/bpf/xdp/     /var/lib/straitgateway/bpf/xdp/
COPY --from=bpf-builder /build/bpf/netkit/  /var/lib/straitgateway/bpf/netkit/
COPY --from=bpf-builder /build/bpf/cgroup/  /var/lib/straitgateway/bpf/cgroup/
COPY --from=bpf-builder /build/bpf/service/ /var/lib/straitgateway/bpf/service/

COPY --from=go-builder /out/sg-agent /usr/local/bin/sg-agent

# ── Volume declarations ──
# /sys/fs/bpf           — BPF FS for pinned maps (Bidirectional)
# /sys/kernel/debug     — BTF / tracing (HostToContainer)
# /lib/modules          — Kernel module loading (read-only)
# /var/run/straitgateway — Runtime sockets and PID file
# /etc/straitgateway    — Config (ConfigMap mount, read-only)
VOLUME ["/sys/fs/bpf", "/sys/kernel/debug", "/lib/modules", "/var/run/straitgateway", "/etc/straitgateway"]

# Metrics / WireGuard / BFD / BGP
EXPOSE 9090/tcp 51820/udp 3784/udp 179/tcp

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/sg-agent", "healthz"] || exit 1

USER root

ENTRYPOINT ["/usr/local/bin/sg-agent"]
CMD ["--config=/etc/straitgateway/config.yaml"]
