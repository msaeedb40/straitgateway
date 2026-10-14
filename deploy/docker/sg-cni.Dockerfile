# ============================================================
# sg-cni — StraitGateway CNI Plugin
# Used as a DaemonSet initContainer installer.
# Copies the CNI binary + conflist to host via hostPath volumes.
# ============================================================

# ── Stage 0: LLVM/eBPF toolchain base ──
FROM golang:1.27.2-bookworm AS bpf-toolchain

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
    -o /out/sg-cni ./cmd/sg-cni

# ── Stage 4: Installer runtime ──
# Uses debian-slim (has /bin/sh) so the initContainer entrypoint script
# can copy the binary AND generate the CNI conflist atomically.
# This image does NOT run as a long-lived daemon.
FROM debian:bookworm-slim AS runtime

LABEL org.opencontainers.image.title="StraitGateway - sg-cni"
LABEL org.opencontainers.image.description="StraitGateway CNI plugin — initContainer installer"
LABEL org.opencontainers.image.source="https://github.com/straitgateway/straitgateway"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="StraitGateway"

COPY --from=go-builder /out/sg-cni /opt/cni/bin/sg-cni
RUN ln -sf /opt/cni/bin/sg-cni /usr/local/bin/sg-cni

# Embed the CNI conflist template — rendered at install time
COPY deploy/cni-config/10-straitgateway.conflist.tmpl \
     /etc/straitgateway/10-straitgateway.conflist.tmpl

# ── Volume declarations ──
# These are the host paths the initContainer writes to.
# The DaemonSet mounts them as hostPath volumes.
# /host/opt/cni/bin    — destination for the CNI binary
# /host/etc/cni/net.d  — destination for the CNI conflist
VOLUME ["/host/opt/cni/bin", "/host/etc/cni/net.d"]

USER root

# The installer script:
#  1. Copies the binary (atomic rename via tmp file)
#  2. Generates the conflist from the embedded template
#  3. Exits 0 — the DaemonSet cni-watcher container keeps it alive after this
ENTRYPOINT ["/bin/sh", "-c", "\
  set -e; \
  echo '[sg-cni] Installing CNI binary...'; \
  cp /opt/cni/bin/sg-cni /host/opt/cni/bin/.sg-cni.tmp; \
  chmod 755 /host/opt/cni/bin/.sg-cni.tmp; \
  mv -f /host/opt/cni/bin/.sg-cni.tmp /host/opt/cni/bin/straitgateway; \
  echo '[sg-cni] Writing CNI conflist...'; \
  cp /etc/straitgateway/10-straitgateway.conflist.tmpl \
     /host/etc/cni/net.d/10-straitgateway.conflist; \
  echo '[sg-cni] Install complete.'; \
"]
