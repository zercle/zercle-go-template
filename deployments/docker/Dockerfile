# Multi-stage build for production-ready Go application
FROM golang AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y git make tini

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimization flags
ARG BUILD_VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" -o service ./cmd/server

# Build arguments: BUILD_VERSION=$(git rev-parse --short HEAD) BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Final stage - distroless minimal runtime
FROM gcr.io/distroless/static

ENV TZ=Asia/Bangkok
ENV LANG=C.UTF-8

WORKDIR /opt/app

# Copy tini static binary from builder
COPY --from=builder /usr/bin/tini-static /usr/bin/tini

# Copy binary from builder
COPY --from=builder /app/service .

# Copy config files
COPY --from=builder /app/configs ./configs

# Expose port
EXPOSE 3000

# Run the application
ENTRYPOINT ["/usr/bin/tini", "--"]

CMD ["service"]

