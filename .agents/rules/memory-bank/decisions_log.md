# Decisions Log

**Last Updated:** 2026-02-21

This document tracks significant architectural and technical decisions made during the project lifecycle. Each decision includes context, alternatives considered, and the rationale for the final choice.

---

## 2026-02: Go Version Upgrade to 1.26

- **Decision:** Upgrade Go version from 1.25.7 to 1.26
- **Context:** Go 1.26 includes performance improvements and new features
- **Impact:** Updated go.mod, Dockerfile, and CI configuration
- **Status:** Completed

### Rationale

- Better garbage collector performance
- Improved type inference
- Enhanced tooling support

---

## 2026-02: Echo v5 Migration

- **Decision:** Migrate from Echo v4 to Echo v5
- **Context:** Echo v5 provides better performance and new features
- **Alternatives Considered:** Keep Echo v4, migrate to Gin
- **Status:** Completed

### Rationale

- Improved routing performance
- Better context handling
- Active maintenance and updates

---

## 2026-02: JWT Cache Implementation

- **Decision:** Add JWT token validation caching
- **Context:** Reduce authentication latency for repeated token validations
- **Alternatives Considered:** No caching, Redis caching
- **Status:** Completed

### Rationale

- In-memory caching provides significant performance gains
- Configurable cache TTL for security
- Simple implementation without external dependencies

---

## 2026-02: Service → Usecase Rename

- **Decision:** Rename all `service/` directories to `usecase/`
- **Rationale:** Align with Clean Architecture terminology and industry standards
- **Impact:** Breaking change for external consumers (if any)
- **Status:** Completed

### Rationale

- "Use case" is more descriptive of business logic
- Aligns with terminology used in DDD and Clean Architecture
- Industry standard in modern Go applications

---

## 2026-02: SQLC Integration

- **Decision:** Use sqlc for type-safe SQL queries
- **Rationale:** Reduce boilerplate, improve type safety, catch errors at compile time
- **Alternatives Considered:** GORM, sqlx, raw SQL
- **Status:** Completed

### Rationale

- Compile-time query validation
- Type-safe database operations
- No runtime reflection overhead
- Clear SQL query visibility

---

## Initial: Argon2id for Password Hashing

- **Decision:** Use Argon2id for password hashing
- **Rationale:** OWASP recommended, memory-hard, resistant to GPU attacks
- **Alternatives Considered:** bcrypt, scrypt, PBKDF2
- **Status:** Implemented with configurable parameters

### Rationale

- OWASP recommendation for password hashing
- Memory-hard to resist GPU/ASIC attacks
- Configurable for different security requirements
- Production defaults: 64MB memory, 3 iterations

---

## Initial: Clean Architecture Implementation

- **Decision:** Follow Clean Architecture with feature-based organization
- **Rationale:** Clear separation of concerns, testability, maintainability
- **Alternatives Considered:** Layered architecture, MVC
- **Status:** Implemented

### Rationale

- Dependencies point inward toward domain
- Business logic decoupled from frameworks
- Easy to test and mock dependencies
- Feature-based organization for colocation

---

## Initial: PostgreSQL with pgx

- **Decision:** Use PostgreSQL with pgx v5 driver
- **Rationale:** High performance, connection pooling, context support
- **Alternatives Considered:** MySQL, SQLite, MongoDB
- **Status:** Implemented

### Rationale

- Pure Go driver (no C dependencies)
- Excellent connection pooling
- Context support for cancellation
- Strong type system alignment

---

## Initial: Viper Configuration

- **Decision:** Use Viper for multi-source configuration
- **Rationale:** Support for env vars, .env files, YAML
- **Alternatives Considered:** envconfig, standard JSON/YAML parsing
- **Status:** Implemented

### Rationale

- Multiple configuration sources
- Environment variable priority
- Type-safe unmarshaling
- Well-maintained and widely used

---

## Initial: Zerolog Logging

- **Decision:** Use Zerolog for structured logging
- **Rationale:** Zero-allocation JSON logging, performance
- **Alternatives Considered:** Zap, logrus, standard logging
- **Status:** Implemented

### Rationale

- Zero-allocation JSON encoding
- Context-aware logging
- Excellent performance
- Structured log format for observability

---

## Initial: Swagger API Documentation

- **Decision:** Use swaggo for Swagger/OpenAPI documentation
- **Rationale:** Auto-generated from code annotations
- **Alternatives Considered:** Manual OpenAPI spec, external tools
- **Status:** Implemented

### Rationale

- Single source of truth in code
- Always in sync with endpoints
- Interactive API exploration
- Standard documentation format

---

## Initial: Docker Multi-stage Build

- **Decision:** Use multi-stage Docker build
- **Rationale:** Minimize production image size
- **Alternatives Considered:** Single-stage build, Distroless
- **Status:** Implemented

### Rationale

- Smaller final image (no build tools)
- Better security posture
- Faster container startup
- Multi-platform support

---

## Reversed Decisions

No reversed decisions to date.

---

## Pending Decisions

### Potential: Redis for Session/Token Storage

- **Status:** Under consideration
- **Context:** Could improve scalability for distributed deployments
- **Alternatives:** In-memory (current), external cache

### Potential: GraphQL API

- **Status:** Under consideration
- **Context:** Could provide more flexible client queries
- **Alternatives:** REST only (current), hybrid

### Potential: Event-Driven Architecture

- **Status:** Under consideration
- **Context:** Could improve decoupling and scalability
- **Alternatives:** Synchronous only (current), message queues

---

## Decision Review Process

1. **Propose** - Document the problem and potential solutions
2. **Evaluate** - Analyze alternatives against requirements
3. **Decide** - Select the best option with rationale
4. **Implement** - Execute the decision
5. **Review** - Assess outcomes and document lessons learned

### Review Checklist

- [ ] Is the decision documented?
- [ ] Are alternatives clearly explained?
- [ ] Is the rationale clear?
- [ ] Are implications understood?
- [ ] Is the impact assessed?
- [ ] Is it communicated to stakeholders?
