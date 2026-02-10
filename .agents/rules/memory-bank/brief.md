# Project Brief

**Last Updated:** 2026-02-10

## Project Identity

- **Name:** zercle-go-template
- **Module:** zercle-go-template
- **Type:** Production-ready Go web application template
- **Go Version:** 1.25.7

## Purpose

This project provides a comprehensive, production-ready Go web application template following Clean Architecture and Hexagonal principles. It serves as a foundation for building scalable, maintainable web services with JWT authentication, user management, and a well-structured codebase.

## Core Requirements

1. **Clean Architecture Implementation**
   - Feature-based organization
   - Clear separation of concerns (Presentation → Business Logic → Infrastructure)
   - Dependency injection for loose coupling

2. **Authentication & Security**
   - JWT-based authentication with access/refresh tokens
   - Argon2id password hashing (OWASP compliant)
   - Secure configuration management

3. **Data Persistence**
   - PostgreSQL database with type-safe queries (sqlc)
   - In-memory repository for testing/development
   - Repository pattern for data access abstraction

4. **API Design**
   - RESTful API with Echo v4 framework
   - Swagger/OpenAPI documentation
   - Standardized response format

5. **Developer Experience**
   - Comprehensive testing (unit, integration, benchmarks)
   - Mock generation with go:generate
   - Structured logging with Zerolog
   - Multi-source configuration (env, .env, YAML)

## Target Audience

- Backend developers building Go web services
- Teams requiring a production-ready starting point
- Projects needing Clean Architecture implementation
- Applications requiring JWT authentication
- Services requiring PostgreSQL integration

## Key Features

- User CRUD operations with validation
- Email/password authentication
- JWT token generation and validation
- Password hashing with Argon2id
- Configurable repository implementations
- Graceful shutdown handling
- Request/response logging middleware
- Health check endpoints
- Swagger API documentation

## Project State

The template is production-ready with:
- Complete implementation of user and auth features
- Dual repository implementations (memory + PostgreSQL)
- Comprehensive test coverage
- Docker support for containerization
- CI/CD configuration (pre-commit hooks, golangci-lint)
- Migration scripts for database schema
