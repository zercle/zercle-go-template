# syntax=docker/dockerfile:1

# -----------------------------------------------------------------------------
# Builder
# -----------------------------------------------------------------------------
FROM golang:1.26 AS builder

WORKDIR /build

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

# -----------------------------------------------------------------------------
# Final
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot /server /server
COPY --from=builder --chown=nonroot:nonroot /build/config.yaml /config.yaml

USER nonroot:nonroot

EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["/server"]
