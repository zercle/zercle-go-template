# Technical Stack & Guidelines

## Core Technologies
- **Language**: Go 1.25+
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL 18+
- **ORM/Query**: `sqlc` (Type-safe SQL), `lib/pq`
- **Config**: Viper
- **Logging**: Zerolog (structured)
- **Error Handling**: `samber/oops`
- **DI**: `samber/do/v2`
- **Validation**: `go-playground/validator`
- **Testing**: `testify`, `go-sqlmock`, `uber-go/mock`
- **Docs**: `swaggo` (Swagger)

## Coding Standards
- **Architecture**: Strict Clean Architecture (Handler -> Service -> Repository -> Domain).
- **Formatting**: Standard `gofmt` / `goimports`.
- **Linting**: `golangci-lint` (v1.64+ recommended).
- **Error Handling**: Explicit verification; use `samber/oops` for rich error context.
- **Logging**: Use `zerolog` for all logs; avoid `fmt.Print`.
- **Naming**: camelCase for local variables, PascalCase for exported.
- **Safety**: Avoid global state; use DI container.

## Database
- Use `uuidv7` for primary keys.
- Migrations handled by `golang-migrate` command via Makefile.
- Queries defined in `sql/queries/*.sql`.

## Testing Rules
- Unit tests for Services and Domain logic.
- Integration tests for Repositories and Handlers.
- Use mocks (`go.uber.org/mock`) for dependencies.
