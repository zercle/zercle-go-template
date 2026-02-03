FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make protoc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /server .
COPY configs/config.yaml ./configs/

EXPOSE 50051

CMD ["./server"]
