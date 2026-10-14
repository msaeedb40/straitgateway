# ============================================================
# sgctl — StraitGateway CLI Tool
# Lightweight CLI, no eBPF deps, runs as non-root
# ============================================================

# ── Stage 1: Go module cache ──
FROM golang:1.27.2-bookworm AS go-deps

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ── Stage 2: Go binary build ──
FROM go-deps AS go-builder

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETARCH=amd64
ARG TARGETOS=linux

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags "-s -w \
      -X 'github.com/straitgateway/straitgateway/pkg/version.Version=${VERSION}' \
      -X 'github.com/straitgateway/straitgateway/pkg/version.GitCommit=${GIT_COMMIT}' \
      -X 'github.com/straitgateway/straitgateway/pkg/version.BuildDate=${BUILD_DATE}'" \
    -o /out/sgctl ./cmd/sgctl

# ── Stage 3: Runtime ──
# distroless/static:nonroot — no shell, fully minimal.
# sgctl makes HTTPS calls to the kube-apiserver; ca-certificates are
# already baked into distroless/static.
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="StraitGateway - sgctl"
LABEL org.opencontainers.image.description="sgctl — StraitGateway CLI for cluster and gateway management"
LABEL org.opencontainers.image.source="https://github.com/straitgateway/straitgateway"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="StraitGateway"

COPY --from=go-builder /out/sgctl /usr/local/bin/sgctl

# ── Volume declarations ──
# ~/.kube/config  — kubeconfig (bind-mount from host when running as a pod or locally)
# /tmp            — temporary output files (e.g. --output=file)
VOLUME ["/root/.kube", "/tmp"]

# UID 65532 = distroless nonroot
USER 65532:65532

ENTRYPOINT ["/usr/local/bin/sgctl"]
