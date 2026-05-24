# AGENTS.md

## Specification Before Implementation

- **Design before you generate.** Clarify objects, collaborations, and boundaries before writing code.
- **Lock intent before you write code.** Make "what we will do / what we won't do" explicit up front.
- **Treat specs as first-class artifacts.** Version-controlled, reviewed, and maintained alongside code.
- **Sync, don't hand off.** Keep specifications and code synchronized — when either side changes, reflect it back. Utilize `.agents/plans/` for plan and progress tracker.

## REASONS Canvas

For non-trivial tasks, structure thinking across these dimensions:

- **R**equirements — What problem are we solving? What is the definition of done?
- **E**ntities — Domain objects (User, Session, Room, Message, etc.) and their relationships.
- **A**pproach — The strategy to meet the requirements (new feature, refactor, optimization, bugfix).
- **S**tructure — Where the change fits in the system: `internal/features/<name>/`, `internal/infrastructure/`, `api/proto/`, `cmd/`, etc.
- **O**perations — Concrete, testable implementation steps.
- **N**orms — Project conventions (Go idioms, package naming, error patterns, DI style).
- **S**afeguards — Non-negotiable constraints (no secrets in code, no panics in request handlers, DB isolation).

## Project Architecture

```
cmd/                      # Entry points (server, client)
├── server/               # gRPC server bootstrap
└── client/               # HTTP client bootstrap
internal/
├── features/             # Feature modules (vertical slices)
│   ├── auth/             # Authentication feature
│   │   ├── di/           # Dependency injection wiring
│   │   ├── domain/       # Domain entities & business rules
│   │   ├── dto/          # Data transfer objects
│   │   ├── handler/      # gRPC & HTTP handlers
│   │   ├── repository/   # Data access (postgres)
│   │   └── service/      # Business logic
│   └── chat/             # Chat feature
├── infrastructure/       # Cross-cutting infrastructure
│   ├── config/           # Configuration (viper)
│   ├── db/               # Database (migrations, queries, postgres client)
│   └── messaging/        # Valkey (Pub/Sub)
└── shared/               # Shared utilities
    ├── di/               # Shared DI container helpers
    ├── errors/           # Domain error sentinels
    ├── middleware/       # HTTP/gRPC middleware
    └── telemetry/       # Logging (zerolog), tracing
api/
├── proto/                # gRPC protobuf definitions
└── openapi/              # REST API spec (Swagger)
pkg/                      # Reusable libraries (uuidgen, etc.)
```

### Feature Layer Pattern

Each feature follows a strict layered architecture:

1. **domain/** — Entities, value objects, repository interfaces (no implementation). Pure business logic with zero external dependencies.
2. **service/** — Business logic orchestration. Depends on domain interfaces, never on infrastructure directly. Defines its own interface for handler consumption.
3. **repository/** — Implementation of domain repository interfaces. Talks to PostgreSQL via pgx/sqlc.
4. **handler/** — gRPC and HTTP handlers. Converts transport-layer requests to service calls. No business logic.
5. **dto/** — Transport-layer request/response structs. Separate from domain entities.
6. **di/** — Wiring via `samber/do` container. Register feature dependencies here.

### Dependency Injection

- Use `github.com/samber/do/v2` for all DI. No package-level global state.
- Feature packages register their providers in `di/provider.go`.
- The `cmd/` entry points build the root container and invoke providers.

## Code Craftsmanship

### Structure for Reuse

- Separate entry-point logic from domain logic. `cmd/` is thin bootstrapping only.
- Return data, not side effects. Handlers return responses; services orchestrate.
- Return errors, don't crash. Every function that can fail returns `error`. Never `panic()` in request handlers.

### Test as You Write

- Name tests as sentences: `TestUser_Validate_EmptyUsername_ReturnsError`.
- Cover happy paths, error paths, and edge cases.
- Tests are living documentation.
- Use `go.uber.org/mock/mockgen` with `//go:generate` directives on interfaces.
- Mocks live in `mock/` sub-packages.

### Design for Reading

- Consistent naming: packages are singular (`service`, `handler`, `repository`, not `services`).
- Extract boilerplate into named helpers.
- Document intent at the component level. Export only what other packages need.

### Make Invalid States Unrepresentable

- Validate at boundaries: domain entities have `Validate()` methods. DTOs are validated in handlers.
- Use constants over magic values (e.g., `domain.StatusOnline`, `domain.StatusOffline`).
- Design types so misuse is hard: use `uuid.UUID`, not `string` for IDs.

### Enrich Errors with Context

- Define named sentinel errors in `internal/shared/errors/`.
- Wrap with `fmt.Errorf("context: %w", err)`, don't flatten to strings.
- Preserve error chains for debugging. Use `errors.Is()` and `errors.As()` for checks.
- Log at service boundaries; return errors to handlers for status code mapping.

### Avoid Mutable Global State

- No package-level mutable variables. Ever.
- Use explicit dependency injection via `samber/do` over global defaults.
- Configuration flows inward from `cmd/` through DI into services.

### Use Concurrency Sparingly

- Only when the problem requires it (e.g., SSE event streaming).
- Keep it localized to one goroutine or feature boundary.
- Ensure all spawned goroutines terminate via `context.Context` cancellation.
- Use `sync.WaitGroup` or `errgroup.Group` for coordinated shutdown.

### Decouple from Environment

- Business logic has no knowledge of env vars, CLI args, or filesystem paths.
- Config structs are defined in `internal/infrastructure/config/`. Services receive only the fields they need.
- Use `spf13/viper` for configuration loading in `cmd/` only.

### Handle Errors Deliberately

- Check every error. Never use `_` for error returns.
- Handle where possible. Retry only for transient failures (network, DB connection).
- Propagate otherwise with context: `fmt.Errorf("failed to create user: %w", err)`.
- Never silently ignore. If an error can't be handled, propagate it up.

### Log Actionable Information Only

- Use `rs/zerolog` for structured logging.
- Log only what someone needs to investigate and fix.
- Structured fields: `log.Error().Err(err).Str("user_id", id.String()).Msg("failed to create user")`.
- Never log secrets (passwords, tokens, keys).
- Use `context.Context` for request-scoped logging fields (trace ID, user ID).

## Security Rules

- Never commit secrets. Use environment variables or `.env` files (gitignored).
- Passwords are hashed with `golang.org/x/crypto/bcrypt`. Never store plaintext.
- JWT secrets must be 256-bit minimum. Rotate in production.
- Validate all input at the handler boundary. Reject unknown fields.
- Use parameterized queries via `pgx`/`sqlc`. Never string-concatenate SQL.
- TLS everywhere in production. gRPC and HTTP both use TLS.

## Testing Standards

- Unit tests: `go test -v -race ./...` — no external dependencies (use mocks).
- Integration tests: `go test -v -tags=integration ./test/integration/...` — requires PostgreSQL and Valkey.
- Coverage: track with `coverage.out`. No hard threshold, but aim to increase.
- Test helpers live in `internal/shared/testutil/` or feature-local `*_test.go` files.

## Tooling

| Tool | Purpose | Command |
|------|---------|---------|
| `mockgen` | Generate mocks from interfaces | `go generate ./...` |
| `sqlc` | Generate type-safe DB queries | `sqlc generate` |
| `golangci-lint` | Linting | `make lint` |
| `golang-migrate` | DB migrations | `make migrate-up / migrate-down` |
| `goimports` | Format imports | `make fmt` |
| `protoc` | Generate gRPC code | See README |

## Iterative Review

- Turn output into a controlled loop, not a one-shot draft.
- For logic corrections: update the spec first, then regenerate code.
- For refactoring: change the code first, then sync back to the spec.
- Verify core functionality before optimizing code quality.
- Make it work, then make it right.
- Use tools effectively.
- Delegate independent tasks to suitable agents if available.
- Always review and correct issues in a timely manner.
