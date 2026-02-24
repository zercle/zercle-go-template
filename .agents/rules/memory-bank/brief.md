# Project Brief

> **⚠️ READ-ONLY FILE** - This file should NOT be modified. It serves as the canonical project overview.

## Project Identity

| Attribute | Value |
|-----------|-------|
| **Name** | zercle-go-template |
| **Type** | Go Chat Application Template |
| **Version** | 1.0.0 |
| **License** | MIT |

## Purpose

Production-ready Go chat application template implementing Clean Architecture with comprehensive testing infrastructure. Designed to accelerate chat application development by providing a solid foundation with best practices built-in, featuring both server and client components with real-time messaging capabilities.

## Core Domain

Chat application with real-time messaging and persistent storage capabilities:

- **Chat Server**: Real-time gRPC-based chat server (cmd/server/main.go)
- **Chat Client**: HTTP chat client (cmd/client/main.go) 
- **Messaging**: Valkey 9 as messaging bus for real-time communication
- **Persistence**: PostgreSQL 18 as persistent data store
- **Authentication**: JWT-based token authentication for secure access

## Target Users

- Developers building real-time chat applications
- Teams needing a scalable messaging platform starting point
- Projects requiring real-time communication with persistent storage

## Key Differentiators

1. **Clean Architecture** - Feature-based organization with clear layer separation
2. **Real-time Messaging** - Valkey 9 powered pub/sub for instant message delivery
3. **Comprehensive Testing** - All layers testable with interface mocking, no external dependencies required
4. **Docker-First** - Multi-stage builds with distroless runtime
5. **Structured Logging** - JSON logging with zerolog

## Technology Stack

| Layer | Technology |
|-------|------------|
| HTTP Framework | Echo v5 |
| Real-time Bus | Valkey 9 (Redis alternative) |
| Database | PostgreSQL 18 with pgx driver |
| SQL Generation | sqlc |
| Configuration | Viper (YAML + env vars) |
| Logging | zerolog |
| Authentication | JWT (golang-jwt/v5) |
| Validation | go-playground/validator |
| Mocking | go.uber.org/mock |

## Quick Start

```bash
# Clone and setup
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template
cp .env.example .env

# Run with Docker (PostgreSQL + Valkey)
docker-compose up -d

# Build and run server
make build-server
./bin/server

# Or build and run client
make build-client
./bin/client
```

## Project State

- ✅ Real-time chat server with WebSocket support
- ✅ Command-line chat client implementation
- ✅ Valkey 9 integration for messaging
- ✅ PostgreSQL 18 for persistent storage
- ✅ JWT authentication system
- ✅ Feature-based architecture with shared models
- ✅ Comprehensive test suite with mocking
- ✅ Docker multi-stage builds

## Constraints

- Go 1.26+ required
- PostgreSQL 18+ and Valkey 9+ for production
- Minimum 256-bit JWT secret
- Non-root container user (security)
- Feature-based organization prevents circular imports
