# Zercle Go Template - Project Brief

## Project Purpose
A production-ready Go RESTful API template that serves as a foundational starting point for building online booking systems and similar web applications. The template eliminates boilerplate setup and provides best practices out of the box.

## Vision
To provide developers with a clean, well-architected Go template that accelerates development while maintaining high standards for code quality, testing, and production readiness. The template should be easily extensible for various business domains.

## Core Objectives
1. **Developer Experience**: Zero to running application in under 5 minutes
2. **Production Ready**: Include all necessary components for deployment (logging, monitoring, health checks, graceful shutdown)
3. **Clean Architecture**: Enforce separation of concerns and maintainability
4. **Testing First**: Comprehensive test coverage with examples
5. **Documentation**: Clear API documentation and code comments

## Success Criteria
- New developers can understand the architecture within 30 minutes
- All tests pass with >80% code coverage
- Application can be deployed to production without modifications
- API is fully documented with Swagger/OpenAPI
- Graceful shutdown works under load

## Constraints
- Go 1.25+ required
- PostgreSQL 18+ as the primary database
- Must support containerized deployment (Podman/Docker)
- Must follow standard Go project layout
- No external service dependencies for core functionality

## Target Users
- Go developers building RESTful APIs
- Teams starting new microservices or monolithic applications
- Developers learning clean architecture in Go
- Organizations needing a standardized Go project template

## Business Domain Focus
Online booking system with:
- User management and authentication
- Service catalog management
- Booking/scheduling functionality
- Payment processing integration
- Availability tracking

## Non-Negotiable Requirements
- Clean architecture with clear layer boundaries
- Type-safe database access using SQLC
- JWT-based authentication
- JSend response format for all APIs
- Structured logging with contextual fields
- Graceful shutdown with 30-second timeout
- Health check endpoints
- Comprehensive test coverage