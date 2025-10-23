# Product Goals & Requirements

## Vision
To be the go-to template for building scalable, high-performance Go microservices at Zercle, ensuring consistency and best practices across projects.

## Core Features
- **Web Framework**: Fiber v2 for high-performance HTTP handling.
- **Database**: PostgreSQL with UUIDv7 for time-sorted, scalable primary keys.
- **Data Access**: Type-safe SQL generation using `sqlc`.
- **Dependency Injection**: Robust DI using `samber/do/v2`.
- **Authentication**: JWT-based secure token management.
- **Observability**: Structured logging with `zerolog` and metrics readiness.
- **Documentation**: Auto-generated Swagger/OpenAPI docs.
- **Standardization**: JSend-compliant JSON responses.

## User Experience
- Developers should find it easy to extend and maintain.
- Clear separation of concerns makes onboarding new developers faster.
- Built-in tools (Makefile, Docker) simplify the dev workflow.
