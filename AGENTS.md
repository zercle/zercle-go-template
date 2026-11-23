# Repository Guidelines

Repo-specific guidance for AI coding agents working in `zercle-go-template`.

`README.md` holds the architecture rationale and the feature-replacement checklist. `Taskfile.yml`, `.golangci.yml`, and `.github/workflows/ci.yml` are the executable source of truth — verify anything here against them.

## Project Overview

An opinionated Go **HTTP service template**: clean (DDD) architecture per feature, `samber/do/v2` DI, echo v5 HTTP, GORM over pgx, a Valkey cache-aside example, zerolog logging, and OpenTelemetry tracing + Prometheus metrics. Four layered features form a **distributed-vending-machines demo** — `catalog` (global product pool), `machines` (vending machines + coin banks), `sales` (purchases), and `reporting` (read-only cross-feature summary) — meant to be copied or replaced; see README §"Adding and deleting features".

Consumers (other Go services) import only `pkg/api/v1` to construct payloads and interpret the `{"error": code, "message": msg}` envelope; `pkg/api/v1` is the canonical source of the wire shapes and stays a leaf package.

## Architecture & Data Flow

Clean (DDD) architecture **per feature**, all dependencies pointing inward. This direction is **enforced by an executable test**, not convention alone (`internal/archguard/`, build tag `unit`, runs in `task test`).

```
consumer services ──> pkg/api/v1    (published wire contract)
handler ──> usecase.Service ──> port.Repository <── infrastructure/postgres/<f>
all layers ──> domain (entities + sentinel errors; standard library only)
infrastructure/* and shared/* ── feature-agnostic, never import features/**
```

A request flows: `cmd/server/main.go` loads config → `internal/app.Build` wires the do injector (`infrastructure: config → telemetry → postgres → valkey → transport`, then the features in list order) → route mounted at `g := e.Group("/api/v1")` → `handler` binds and validates the `pkg/api/v1` request, translates it into domain inputs, calls `usecase.Service` → the usecase (speaking domain types only) parses ids and applies business rules → `port.Repository` (GORM, in `internal/infrastructure/postgres/<f>`) behind an optional `CachedRepository` decorator (Valkey cache-aside) → the handler maps the domain result back to the `pkg/api/v1` response; domain sentinel errors are mapped to the HTTP envelope by `shared/errors`.

**Dependency gates** — `internal/archguard/` parses imports (`parser.ImportsOnly`, skipping `mock/` and `_test.go`); `layering_test.go` holds the direction rules, `imports_test.go` the import-surface rules. A violation fails with `package %q violates %s`. If a change trips a rule, **restructure the change — never weaken the rule**:

| Rule | Forbids |
|---|---|
| `domain-is-innermost` | anything non-stdlib in `features/<f>/domain` |
| `usecase-depends-on-domain-and-port` | the `usecase` package importing anything outside its own domain/port |
| `port-depends-only-on-domain` | the `port` package referencing anything but its own `domain` |
| `handler-ignores-port-and-adapters` | the `handler` package importing `port` or the postgres driven adapters |
| `driven-adapter-ignores-usecase-and-handler` | `infrastructure/postgres/<f>/**` importing `usecase` or `handler` |
| `infrastructure-is-feature-agnostic` | `infrastructure/**` importing `features/**` (except the `postgres/<f>` driven adapters) |
| `features-do-not-import-other-features` | one feature importing any other feature's packages |
| `shared-ignores-features` | `shared/**` importing `features/**` |
| `published-contract-is-leaf` | `pkg/api/**` importing anything but stdlib and `pkg/api/errcodes` |
| `composition-root-is-top` | anything under `internal/` importing `internal/app` |

## Key Directories

- `cmd/server/` — thin entry point: loads config, sets `Version`/`CommitSHA`/`BuildTime` (ldflags), installs the signal handler, calls `app.Run`.
- `cmd/migrate/` — self-contained migration runner (`up`/`down [N]`/`force`/`version`); embeds the root `migrations/` directory via `go:embed` and applies it with golang-migrate's `iofs` source.
- `internal/app/` — the composition root (`Build`, `Run`); the only place wiring order is defined. Holds one wiring file per feature (`catalog.go`, …) plus the ordered feature list in `features.go`.
- `internal/archguard/` — the dependency gates above (`layering_test.go`, `imports_test.go`).
- `internal/features/{catalog,machines,sales,reporting}/` — the demo features. Layers, each its own package: `domain` (entities + sentinels), `port` (outbound iface + `mock/`), `usecase` (`Service` iface + `Usecase`, one file per operation, domain types only + `mock/`), `handler` (echo v5 driving adapter). Features never import each other; `sales` reads catalog/machines data through its own port in one transaction.
- `internal/infrastructure/` — cross-cutting technical plumbing: `config`, `postgres` (the postgres client + one `<f>/` driven adapter per feature with its GORM repository and `models/`), `valkey`, `lifecycle`, `telemetry` (the logger, the tracer provider, the propagator, the Prometheus registry, and the health registry), `transport` (router + server lifecycle, with `transport/middleware/` for request-id, request-logger, secure, prometheus, cors, otel).
- `internal/shared/` — cross-feature code: `errors/` (typed boundary errors + HTTP mapper).
- `internal/testutil/` — shared test helpers + `fixtures/`.
- `migrations/` — versioned SQL for the whole service (one directory, one namespace).
- `pkg/api/errcodes/` — published error-code constants. `pkg/api/v1/` — the published wire contract (canonical request/response types, leaf package).
- `test/e2e/` — end-to-end tests.

## Development Commands

[Task](https://taskfile.dev) is the runner — prefer `task <name>` over raw `go` (it encodes flags you will get wrong by hand):

- `task build` / `task run` — build `bin/server` with version ldflags / run it (`run` depends on `build`).
- `task test` (alias `task test-unit`) — unit suite with coverage profile.
- `task test-integration` — needs live postgres + Valkey (`docker compose up -d postgres valkey`, then `cp .env.example .env`).
- `task test-e2e` — boots the full server (skips if the DB/Valkey TCP probe fails).
- `task lint` / `task fmt` (gofumpt + goimports) / `task tidy` / `task verify` (tidy + `git diff --exit-code go.mod go.sum`).
- `task generate` — regenerate mockgen mocks (`go generate ./...`); run after touching a port or usecase interface.
- `task migrate-up` / `migrate-down [N=1]` — run the self-contained runner (`go run ./cmd/migrate up|down N`), which applies the root `migrations/` directory; `migrate-create NAME=<name>` needs the golang-migrate CLI on `PATH` and writes into `migrations/` (no `FEATURE` variable).

**Build tags are mandatory; plain `go test ./...` runs zero tests** — every `*_test.go` is gated by `unit`, `integration`, or `e2e`:

```bash
go test -race -tags=unit -run TestName ./internal/features/catalog/usecase/...
go test -race -tags=integration ./internal/infrastructure/postgres/catalog/...
go test -race -tags=unit ./internal/archguard/...   # dependency gates only
```

The Taskfile's `dotenv` is deliberately **per-task** (`run`/`test-integration`/`test-e2e`/`migrate-*` only). Do **not** make it global: viper binds env over `config.yaml`, so a global dotenv makes `task test` assert wrong values.

## Code Conventions & Common Patterns

- **Naming.** Layer packages are fixed lowercase nouns (`domain`, `port`, `usecase`, `handler`). Interfaces are named by role, not feature: `usecase.Service` (inbound), `port.Repository` (outbound). Impls: `Usecase`, `Repository`, `CachedRepository` (decorator), `Handler`, `Application`. Constructors are `New*` returning concrete types; each feature's wiring entrypoint is `register<Feature>(c do.Injector) error` in `internal/app/<f>.go`.
- **DI (samber/do/v2).** `internal/app.Build` is the composition root in fixed order; each feature wires itself via `do.Provide(...)` in its `internal/app/<f>.go` file. Providers implementing `Shutdowner*` are shut down by the container automatically — do not additionally close them in `Application.shutdown`.
- **Error handling.** Three tiers: (1) **domain sentinels** — package-level `Err<Name>` via `errors.New` in `features/*/domain/errors.go`; (2) **boundary sentinels** — `*AppError{Code, Message, HTTPStatus, Cause}` in `shared/errors` (`ErrNotFound`, `ErrInvalidInput`, `ErrInternal`, …), codes from `pkg/api/errcodes` so served and published codes cannot drift; (3) **registration** — features call `apperrors.RegisterSentinel(domain.ErrX, apperrors.ErrY)` in `internal/app/<f>.go`, and the handler maps any error via `status, body := apperrors.HTTPError(err); return c.JSON(status, body)`. Framework errors (404/405/413) go through the same mapper. The envelope is always `{"error": code, "message": msg}`.
- **Config.** Loaded from `config.yaml` + unprefixed env vars (viper + validator); every leaf is explicitly bound in `internal/infrastructure/config`. `CONFIG_FILE` overrides the path. `app.Build` calls `cfg.Validate()`, so bad values fail startup. `CATALOG_ENABLED` / `MACHINES_ENABLED` / `SALES_ENABLED` / `REPORTING_ENABLED` gate each demo feature's providers and routes entirely. `HTTP_HSTS_MAX_AGE` (duration, default `0s`) controls the `Strict-Transport-Security` header.
- **Persistence.** The SQL schema is owned by golang-migrate files in the root `migrations/` directory (embedded via `//go:embed`). GORM model tags only map to existing columns — **`AutoMigrate` is never used**.
- **Generated code — regenerate, never hand-edit.** `//go:generate go tool mockgen` directives on `usecase/service.go` and `port/repository.go` write into `usecase/mock/` and `port/mock/`. Tests won't compile after an interface change until `go generate ./...`. `*/mock/` is excluded from lint.
- **Echo v5 gotcha.** Handlers take `*echo.Context` (v5 changed Context from interface to struct) and return `error`. This is correct, not a typo.
- **HTTP middleware.** The stack is echo's built-ins (`RecoverWithConfig`, `RequestIDWithConfig`, `RequestLoggerWithConfig`, `SecureWithConfig`, `CORSWithConfig`) plus `github.com/labstack/echo-otel/v5` (tracing) and `github.com/labstack/echo-prometheus` (metrics). The `middleware` package only adapts them: zerolog output for RequestLogger, a UUID v7 generator and client-id validation for RequestID, config-driven HSTS for Secure, the repository's own Prometheus registry for `Prometheus`/`MetricsHandler`, and the tracer provider + propagator (with a no-op metrics recorder) for OTel. The registration order in `http.go` is deliberate (RequestID outermost, OTel outside Recover, RequestLogger outside Recover, Prometheus before Recover). A recovered panic must keep returning the shared `{"error", "message"}` envelope, and the stack is logged as a structured `stack` field.
- Formatting: gofumpt + goimports (`.editorconfig`: tabs for `.go`, 2-space YAML, LF, final newline). `gocyclo` max-complexity is 15; CI fails on `gofmt -s` diffs.

## Important Files

- `cmd/server/main.go` — process entry point (config, ldflags vars, signals, `app.Run`).
- `internal/app/app.go` — composition root: DI wiring order, `Build`, `Run`.
- `internal/app/features.go` — the ordered feature list; `internal/app/catalog.go` — the canonical feature wiring (flag gate, sentinel registration, providers, route mount).
- `internal/archguard/layering_test.go` — the dependency-direction gates (source of truth for layering); `imports_test.go` holds the import-surface gates.
- `internal/infrastructure/config/config.go` — config struct, per-leaf env binding, `Validate`.
- `internal/shared/errors/` — `app_error.go`, `sentinel.go`, `mapper.go`, `status.go` (error→HTTP mapping).
- `internal/infrastructure/transport/` — `http.go` (echo bootstrap, error handler) and `shutdown.go` (graceful shutdown), with `middleware/` for request-id, request-logger, secure, prometheus, cors, otel.
- `internal/infrastructure/valkey/` — `client.go` (`NewClient`, `NewCacheAside`), `health.go`.
- `cmd/migrate/main.go` — embeds `migrations/` via `go:embed` and applies it with golang-migrate's `iofs` source.
- `Taskfile.yml`, `.golangci.yml`, `config.yaml`, `.env.example`, `compose.yml`, `Containerfile`, `.github/workflows/ci.yml`.

## Runtime/Tooling Preferences

- **Go 1.27** (`go.mod` directive is the effective floor; CI uses `actions/setup-go` with `go-version: stable`). Module: `github.com/zercle/zercle-go-template`. stdlib `uuid` is used directly (no third-party UUID lib).
- **Task** (taskfile.dev) is the command runner; **golang-migrate CLI** is needed only for the `task migrate-*` variants.
- Key deps: `labstack/echo/v5`, `labstack/echo-otel/v5` (tracing), `labstack/echo-prometheus` (metrics), `gorm.io/gorm` + `gorm.io/driver/postgres`, `valkey-io/valkey-go` (+ `valkeyaside`), `samber/do/v2`, `spf13/viper`, `go-playground/validator/v10`, `rs/zerolog`, `golang-migrate/migrate/v4`, OTel `otel`/`trace`/`sdk` (+ `otlptracehttp`) for the OTLP trace export. Test stack: `stretchr/testify`, `go.uber.org/mock`, `DATA-DOG/go-sqlmock`.
- `mockgen` is declared as a Go **tool** dependency (`tool go.uber.org/mock/mockgen`) — invoked via `go generate`, not a global binary.
- Lint: **golangci-lint v2** (`.golangci.yml`) with `gocyclo` min-complexity 15, `errcheck.check-type-assertions: true`, and `wrapcheck`. Run `task fmt` and `task lint` before committing.
- Local services: `docker compose up -d postgres valkey` (`compose.yml` uses `postgres:18-alpine` and `valkey:9-alpine`). `Containerfile` is a two-stage distroless/non-root server image.

## Testing & QA

- **Frameworks.** testify (`require` for preconditions, `assert` for final checks; integration suites use `suite.Suite`); `go.uber.org/mock` generated mocks; `go-sqlmock` for GORM repository unit tests; `httptest` + `e.ServeHTTP` for handler tests. Most unit tests call `t.Parallel()`. Test names follow `Test<Type>_<Method>_<Case>`.
- **Tags.** `unit` (hermetic, mocked — 30+ files), `integration` (live postgres + Valkey, under `internal/infrastructure/postgres/<f>/`), `e2e` (`test/e2e/server_e2e_test.go`, boots the server, skips when infra is unreachable). **No tag = no test compiled.**
- **Architecture gate.** `go test -tags=unit ./internal/archguard/...` — run this after any refactor that moves imports or packages.
- **Coverage.** 60% total gate, enforced in CI (`ci.yml` unit job) — not in the Taskfile. Measure locally with `task test-unit` then `go tool cover -func=coverage.out | grep total`.
- **Integration setup.** No testcontainers: tests read `config.Load()` (env-bound), apply migrations in-suite via the embedded `iofs` FS, and `TRUNCATE` between cases; they hard-fail (not skip) when infra is missing, with a production guard.
- **CI.** `.github/workflows/ci.yml`: `lint` → `unit` (60% gate, codecov, coverage artifact) → `integration` (postgres + Valkey service containers) → `build`; **lint fails on `gofmt -s` diffs and go.mod/go.sum drift**. `security.yml` runs a weekly Trivy + govulncheck scan; `cd.yml` publishes a multi-arch image on a `v*` tag.

## Gotchas

- `cp .env.example .env` (needed for integration/e2e) combined with a **global** dotenv would silently override `task test`'s asserted config values — hence per-task dotenv.
- Changing a port or usecase interface without `task generate` leaves a stale mock and a non-compiling tree.
- The `catalog`, `machines`, `sales`, and `reporting` features are a replaceable demo: adding a feature is its directory, one entry in the feature list in `internal/app/features.go`, its wiring file `internal/app/<f>.go`, and (if it owns schema) SQL files in `migrations/`; replacing them is the reverse (feature dirs, feature-list entries, wiring files, `pkg/api/v1` types, `config.yaml`/`.env.example`).
- The request id is a UUID v7 (stdlib `uuid.NewV7()`), not a random string; client-supplied ids are validated against `[A-Za-z0-9_-]` with a 128-char cap and replaced if invalid.
- The HTTP metrics are Prometheus-native (from `echo-prometheus`, not the OTel semantic-convention series), and their `url` label is the route template — renaming a route changes the label, so old series do not merge with new ones.
