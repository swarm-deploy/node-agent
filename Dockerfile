# syntax=docker/dockerfile:1

FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/node-agent ./cmd/node-agent

FROM alpine:3.21.7

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/node-agent /usr/local/bin/node-agent

EXPOSE 9000

LABEL org.opencontainers.image.title="node-agent"
LABEL org.opencontainers.image.description="Agent for Swarm nodes"
LABEL org.opencontainers.image.url="https://github.com/swarm-deploy/node-agent"
LABEL org.opencontainers.image.source="https://github.com/swarm-deploy/node-agent"
LABEL org.opencontainers.image.vendor="swarm-deploy"
LABEL org.opencontainers.image.version="$APP_VERSION"
LABEL org.opencontainers.image.created="$BUILD_TIME"
LABEL org.opencontainers.image.licenses="Apache 2.0"

ENTRYPOINT ["/usr/local/bin/node-agent"]
