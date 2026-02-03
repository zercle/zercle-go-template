# Multi-stage Containerfile for zercle-go-template
# Optimized for Podman with Docker CLI compatibility

# =============================================================================
# Builder stage
# =============================================================================
FROM mirror.gcr.io/golang:alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
      -X main.Version=${VERSION} \
      -X main.CommitSHA=${COMMIT_SHA} \
      -X main.BuildTime=${BUILD_TIME}" \
    -o /server ./cmd/server

# =============================================================================
# Runtime stage
# =============================================================================
FROM mirror.gcr.io/alpine:latest

RUN apk add --no-cache ca-certificates tzdata curl git

RUN addgroup -g 65534 -S appgroup && \
    adduser -u 65534 -S appuser -G appgroup

COPY --from=builder /server /usr/local/bin/server
COPY --from=builder /build/internal/infrastructure/database/migrate/migrations /migrations

RUN chown -R appuser:appgroup /usr/local/bin/server /migrations

USER appuser:appgroup

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["server"]