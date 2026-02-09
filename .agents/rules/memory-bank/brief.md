# Project Brief: Zercle Go Template

## Project Overview

**Name:** Zercle Go Template  
**Type:** Production-Ready REST API Template  
**Language:** Go (Golang)  
**Version:** Current (active development)  

The Zercle Go Template is a comprehensive, production-ready starting point for building scalable, maintainable REST APIs in Go. It implements industry best practices including clean architecture, domain-driven design principles, and a robust testing framework. The template eliminates boilerplate setup and provides a solid foundation for API development.

## Core Requirements

### Functional Requirements
- RESTful API endpoints for user management (CRUD operations)
- JWT-based authentication and authorization system
- Secure password hashing (bcrypt)
- Database migrations and schema management
- API documentation via Swagger/OpenAPI specification
- Structured logging with configurable levels
- Environment-based configuration management
- Comprehensive error handling and validation

### Non-Functional Requirements
- High code quality with linting (golangci-lint)
- Unit and integration test coverage
- Docker containerization for consistent deployments
- Pre-commit hooks for code quality enforcement
- Clean, maintainable codebase with clear separation of concerns
- Type-safe database queries using SQLC
- Efficient performance with connection pooling

## Goals

### Primary Goals
1. **Accelerate Development:** Provide a ready-to-use foundation that eliminates repetitive setup tasks, allowing developers to focus on business logic from day one.
2. **Ensure Best Practices:** Embed industry-standard patterns, conventions, and architectural principles to guide developers toward maintainable, scalable code.
3. **Reduce Technical Debt:** Establish a solid foundation with proper testing, error handling, and code organization that prevents accumulation of technical debt.
4. **Enable Rapid Prototyping:** Support quick iteration and experimentation while maintaining production-grade quality standards.

### Secondary Goals
1. **Educational Resource:** Serve as a reference implementation for Go API development, demonstrating proper use of frameworks and libraries.
2. **Consistency:** Provide a standardized approach across multiple projects within the organization, reducing context switching.
3. **Team Onboarding:** Lower the barrier to entry for new team members by providing a familiar, well-documented structure.
4. **Scalability:** Support horizontal scaling through stateless design and proper separation of concerns.

## Key Capabilities

### Authentication & Authorization
- JWT token generation and validation
- Middleware-based route protection
- Secure password hashing with bcrypt
- Token refresh mechanism
- Role-based access control (RBAC) foundation

### User Management
- Complete CRUD operations for user entities
- Email uniqueness validation
- Password update functionality
- User profile management
- Soft delete support (extensible)

### Architecture & Design
- **Clean Architecture:** Layered structure with clear boundaries (handler → usecase → repository → domain)
- **Domain-Driven Design:** Feature-based organization with domain models at the core
- **Dependency Injection:** Container-based dependency management
- **Repository Pattern:** Abstract data access with multiple implementations (SQL, in-memory)
- **Use Case Pattern:** Business logic encapsulation independent of frameworks

### Database Integration
- PostgreSQL as primary database
- SQLC for type-safe SQL query generation
- Database migrations with version control
- Connection pooling and management
- Transaction support
- Integration test database via Docker Compose

### API Documentation
- Swagger/OpenAPI 2.0 specification
- Interactive API documentation UI
- Auto-generated documentation from code annotations
- Request/response schema definitions

### Development Tools
- **Makefile:** Common development tasks (build, test, run, lint, docker)
- **Pre-commit Hooks:** Automated code quality checks
- **Golangci-lint:** Comprehensive static analysis
- **Docker:** Multi-stage builds for optimized images
- **Docker Compose:** Local development environment with database

### Testing Infrastructure
- Unit tests with mocking (gomock)
- Integration tests with test database
- Table-driven test patterns
- Test coverage reporting
- Isolated test environments

### Configuration Management
- YAML-based configuration files
- Environment variable overrides
- Viper for configuration loading
- Type-safe configuration structs
- Validation of configuration values

### Logging & Monitoring
- Structured logging with Zerolog
- Configurable log levels (debug, info, warn, error)
- Request/response logging middleware
- Error stack trace capture
- Contextual logging with correlation IDs

### Error Handling
- Custom error types with context
- Consistent error response format
- HTTP status code mapping
- Error wrapping and unwrapping
- Recovery middleware for panic handling

### Security Features
- Input validation and sanitization
- SQL injection prevention (via SQLC)
- CORS support
- Rate limiting foundation
- Secure headers middleware

## Technology Stack

### Core Framework
- **Echo v4:** High-performance, minimalist web framework
- **Go 1.x:** Modern Go language features and standard library

### Database & ORM
- **PostgreSQL:** Relational database
- **SQLC:** Type-safe SQL query generator
- **pgx:** PostgreSQL driver with connection pooling

### Authentication
- **JWT (golang-jwt/jwt):** Token-based authentication
- **bcrypt:** Password hashing

### Configuration
- **Viper:** Configuration management
- **YAML:** Configuration file format

### Logging
- **Zerolog:** Zero-allocation JSON logger

### Documentation
- **Swagger/OpenAPI:** API specification
- **swaggo:** Swagger code generation

### Development Tools
- **golangci-lint:** Linting and static analysis
- **gomock:** Mock generation for testing
- **pre-commit:** Git hook management
- **Docker:** Containerization
- **Make:** Build automation

## Project Structure

```
zercle-go-template/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── container/       # Dependency injection container
│   ├── errors/         # Custom error types
│   ├── feature/        # Feature modules (auth, user, etc.)
│   ├── infrastructure/  # External dependencies (db, logger)
│   └── middleware/      # HTTP middleware
├── api/docs/           # Swagger documentation
├── configs/            # Configuration files
├── .agents/rules/memory-bank/  # Memory bank system
└── plans/              # Architecture and migration plans
```

## Current State

**Status:** Active Development  
**Stability:** Production-Ready Template  
**Documentation:** In Progress  
**Test Coverage:** Comprehensive (unit + integration)

The template is actively maintained and serves as the foundation for multiple production APIs. All core features are implemented and tested, with room for extension based on specific project requirements.

## Constraints & Considerations

- **Read-Only Brief:** This file serves as the foundation document and should not be modified without explicit approval
- **Go Version:** Requires Go 1.21+ for modern language features
- **Database:** PostgreSQL is the supported database (extensible to others)
- **Deployment:** Designed for containerized deployments (Docker/Kubernetes)
- **Stateless:** Application is stateless; session data stored in JWT tokens
- **Testing:** Requires Docker for integration test database

---

**Last Updated:** 2026-02-08  
**Maintained By:** Zercle Development Team  
