FROM golang:1.26-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /migrate ./cmd/migrate

FROM alpine:latest

RUN apk add --no-cache ca-certificates

RUN addgroup -g 65534 -S appgroup && \
    adduser -u 65534 -S appuser -G appgroup

COPY --from=builder /migrate /usr/local/bin/migrate
COPY --from=builder /build/internal/infrastructure/db/migrations /migrations
COPY --from=builder /build/configs/config.yaml /etc/zercle/config.yaml

RUN chown -R appuser:appgroup /usr/local/bin/migrate /migrations /etc/zercle

USER appuser:appgroup

WORKDIR /etc/zercle

ENTRYPOINT ["migrate"]