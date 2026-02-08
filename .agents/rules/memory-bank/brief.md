# Zercle Go Template - Project Brief

## Project Overview

**Zercle Go Template** is a production-ready Go backend template designed for rapid REST API development. It implements Clean Architecture principles with domain-driven design, providing a solid foundation for building scalable and maintainable web applications.

## What We're Building

A comprehensive Go backend template that includes:

- **REST API Server**: Built with Echo v4 framework for high-performance HTTP routing
- **Authentication System**: JWT-based authentication with secure token generation and validation
- **User Management**: Complete CRUD operations for user entities with validation
- **Type-Safe Database Access**: SQL-generated code using sqlc for compile-time type safety
- **Production Tooling**: Docker support, comprehensive testing, linting, and pre-commit hooks

## Core Requirements

### Architecture
- Clean Architecture with clear separation of concerns
- Domain-driven design with feature-based organization
- Repository pattern for data access abstraction
- Dependency injection for testable components

### Features
- JWT authentication (login, token refresh, middleware protection)
- User CRUD operations with validation
- Password hashing with bcrypt
- Structured logging with zerolog
- Configuration management with Viper
- Swagger/OpenAPI documentation

### Infrastructure
- PostgreSQL database with pgx/v5 driver
- Docker support for containerization
- Database migrations
- Type-safe SQL queries via sqlc

### Quality Assurance
- Unit tests with table-driven test patterns
- Integration tests for database operations
- Mock generation with go.uber.org/mock
- Linting with golangci-lint
- Pre-commit hooks for code quality

## Goals

1. **Provide Solid Foundation**: Offer a production-ready starting point for new Go projects, eliminating boilerplate setup and ensuring best practices from day one.

2. **Demonstrate Best Practices**: Showcase idiomatic Go patterns, Clean Architecture implementation, and industry-standard practices for building backend services.

3. **Enable Rapid Development**: Reduce time-to-market by providing pre-built authentication, user management, and infrastructure components that can be extended quickly.

4. **Support Multiple Environments**: Enable seamless transitions between development, staging, and production with environment-aware configuration and logging.

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.25.7 |
| Web Framework | Echo v4 |
| Database | PostgreSQL with pgx/v5 |
| SQL Generation | sqlc |
| Authentication | JWT (golang-jwt/jwt/v5) |
| Configuration | Viper |
| Logging | zerolog |
| Validation | go-playground/validator/v10 |
| Documentation | Swagger/OpenAPI (swaggo) |
| Testing | testify, go.uber.org/mock |
| Linting | golangci-lint |
| Containerization | Docker |

## Project Structure

```
zercle-go-template/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── container/        # Dependency injection container
│   ├── errors/           # Custom error types
│   ├── feature/          # Feature modules (user, auth, etc.)
│   ├── infrastructure/  # Database, external services
│   ├── logger/           # Logging setup
│   └── middleware/       # HTTP middleware
├── api/docs/            # Swagger documentation
├── configs/             # Configuration files
└── .agents/rules/memory-bank/  # Memory Bank documentation
```

## Key Design Decisions

- **Feature-Based Organization**: Each feature (user, auth) is self-contained with domain, handler, repository, and usecase layers
- **Domain Independence**: Domain layer has no external dependencies, ensuring pure business logic
- **Type Safety**: sqlc generates type-safe Go code from SQL queries, eliminating runtime SQL errors
- **Zero-Dependency Domain**: Domain entities and business rules are isolated from frameworks and databases
- **Mock-Friendly Design**: Interfaces defined at usecase and repository layers enable easy testing with mocks
