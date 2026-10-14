# ============================================================
# sg-controller — StraitGateway Kubernetes Controller
# Standard Deployment — no privileged access needed
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
    -o /out/sg-controller ./cmd/sg-controller

# ── Stage 3: Runtime ──
# distroless/static:nonroot — no shell, no package manager, minimal CVE surface.
# Controller only needs TLS connectivity to kube-apiserver (ca-certs are baked in).
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="StraitGateway - sg-controller"
LABEL org.opencontainers.image.description="StraitGateway Kubernetes controller-manager (Gateway API, CRDs)"
LABEL org.opencontainers.image.source="https://github.com/straitgateway/straitgateway"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="StraitGateway"

COPY --from=go-builder /out/sg-controller /usr/local/bin/sg-controller

# ── Volume declarations ──
# /tmp          — Leader-election lease scratch (in-memory or emptyDir in pod)
# /var/run/secrets — Projected SA token (automatically mounted by K8s)
VOLUME ["/tmp"]

# Metrics and webhook port
EXPOSE 9090/tcp 9443/tcp

HEALTHCHECK --interval=20s --timeout=5s --start-period=15s --retries=3 \
    CMD ["/usr/local/bin/sg-controller", "healthz"] || exit 1

# UID 65532 = distroless nonroot
USER 65532:65532

ENTRYPOINT ["/usr/local/bin/sg-controller"]
CMD ["--leader-elect=true", "--sync-period=30s"]
