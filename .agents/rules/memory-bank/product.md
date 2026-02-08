# Zercle Go Template - Product Documentation

## Purpose

Zercle Go Template solves the **boilerplate problem** for Go backend development. Every new Go project requires setting up the same foundational components: HTTP routing, authentication, database connections, configuration management, and testing infrastructure. This template eliminates that repetitive setup work, allowing developers to focus on business logic from day one.

## Target Users

### Primary Audience
- **Backend developers** building REST APIs in Go
- **Engineering teams** establishing new microservices
- **Technical leads** standardizing project structures
- **Developers learning** Clean Architecture in Go

### Use Cases
1. **Startup MVPs**: Rapidly prototype with production-ready foundations
2. **Enterprise microservices**: Consistent patterns across service boundaries
3. **Educational projects**: Learn Clean Architecture with working examples
4. **Hackathon submissions**: Skip setup, focus on features

## Key Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **JWT Authentication** | Complete auth flow with access/refresh tokens, secure middleware |
| **User Management** | Full CRUD with validation, password hashing, profile operations |
| **Type-Safe Database** | sqlc-generated code eliminates SQL injection and runtime errors |
| **Clean Architecture** | Clear separation: Domain → Use Case → Repository → Handler |
| **Dual Repository Support** | In-memory for testing, PostgreSQL for production |
| **Structured Logging** | Zerolog with configurable levels and formats |
| **Configuration Management** | Multi-layer config: env vars → .env → YAML → defaults |
| **API Documentation** | Auto-generated Swagger/OpenAPI specs |

### Developer Experience

- **Zero-config startup**: Works out of the box with sensible defaults
- **Hot reload**: Docker Compose for development
- **Pre-commit hooks**: Automated linting and formatting
- **Comprehensive tests**: Unit and integration test patterns
- **Makefile commands**: Common tasks abstracted to simple commands

## UX Goals

### Easy to Use
- Single command to start: `make docker-up`
- Clear directory structure following Go conventions
- Extensive inline code documentation
- Working examples for all patterns

### Well-Documented
- README with setup instructions
- Swagger UI at `/swagger/index.html`
- Architecture decision records
- This Memory Bank for AI context

### Extensible
- Feature-based organization: add new features by copying `user/` pattern
- Interface-based design: swap implementations without changing business logic
- Configuration-driven: behavior changes via config, not code

## Roadmap Suggestions

### Phase 1: Foundation (Current)
- [x] Clean Architecture structure
- [x] User authentication
- [x] PostgreSQL integration
- [x] Docker containerization
- [x] Testing infrastructure

### Phase 2: Enhanced Security
- [ ] Rate limiting middleware
- [ ] CORS configuration
- [ ] Request validation middleware
- [ ] Audit logging
- [ ] Role-based access control (RBAC)

### Phase 3: Operational Excellence
- [ ] Health check endpoints
- [ ] Metrics endpoint (Prometheus)
- [ ] Distributed tracing
- [ ] Graceful shutdown handling
- [ ] Request ID propagation

### Phase 4: Developer Tools
- [ ] Code generation for new features
- [ ] Database seeding scripts
- [ ] Migration rollback utilities
- [ ] Load testing setup (k6)
- [ ] CI/CD pipeline examples

### Phase 5: Advanced Features
- [ ] WebSocket support
- [ ] Background job processing
- [ ] File upload handling
- [ ] Multi-tenancy support
- [ ] API versioning strategy

## Acceptance Criteria

A project built from this template should:

1. **Compile without warnings** at `go build ./...`
2. **Pass all tests** with `go test ./...`
3. **Pass linting** with `golangci-lint run`
4. **Start successfully** with `make docker-up`
5. **Serve requests** at `http://localhost:8080`
6. **Show Swagger UI** at `http://localhost:8080/swagger/index.html`
7. **Authenticate users** via JWT tokens
8. **Perform CRUD** on user resources
9. **Log structured output** to stdout
10. **Connect to PostgreSQL** for data persistence

## Success Metrics

- New feature can be added in under 30 minutes
- 100% test coverage for domain layer
- Zero security vulnerabilities in dependency scan
- p99 latency < 10ms for simple requests
- Memory usage < 100MB at idle
