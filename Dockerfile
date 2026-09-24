# syntax=docker/dockerfile:1
FROM alpine:3.20

ARG TARGETARCH
ARG BINARY=straitd

RUN apk add --no-cache ca-certificates iproute2 bash

# Install all binaries from local dist or build context
COPY dist/linux-${TARGETARCH}/ /usr/local/bin/

# Copy CNI conflist
RUN mkdir -p /etc/straitgateway
COPY cni/conflist/straitgateway.conflist /etc/straitgateway/straitgateway.conflist

ENV BINARY_NAME=${BINARY}

# Symlink targeted binary to /entrypoint
RUN ln -sf /usr/local/bin/${BINARY} /entrypoint

ENTRYPOINT ["/entrypoint"]
