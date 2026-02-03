FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./cmd/client

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /client .
COPY configs/config.yaml ./configs/

EXPOSE 8080

CMD ["./client"]
