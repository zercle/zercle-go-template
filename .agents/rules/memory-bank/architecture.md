# System Architecture

## High-Level Overview
The application follows **Clean Architecture** combined with **Domain-Driven Design (DDD)**. Dependencies point inwards.

## Layers
1.  **Domain (`internal/core/domain`)**:
    *   Entities, Value Objects, and Domain Errors.
    *   Pure Go code, no external dependencies.

2.  **Ports (`internal/core/port`)**:
    *   Interfaces defining input (Service) and output (Repository/External) boundaries.

3.  **Service (`internal/core/service`)**:
    *   Business logic implementation.
    *   Depends on Domain and Ports.

4.  **Adapters (`internal/adapter`)**:
    *   **Handler (`internal/adapter/handler`)**: HTTP entry points (Fiber controllers). Maps DTOs to Domain models.
    *   **Storage (`internal/adapter/storage`)**: Database implementations (Repository interfaces). Uses `sqlc`.

5.  **Infrastructure (`internal/infrastructure`)**:
    *   Platform-specific code (Config, Server, DI Container).
    *   `container`: Wires everything together using `samber/do`.

## Data Flow
Request -> Middleware -> Handler -> Service -> Repository -> Database
Response <- Middleware <- Handler <- Service <- Repository <- Database

## Key Modules
- **Auth**: JWT handling and middleware.
- **Server**: Fiber app configuration and lifecycle.
- **Logger**: Centralized `zerolog` configuration.
