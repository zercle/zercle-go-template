# Repository Guidelines

Repo-specific guidance for AI coding agents working in `zercle-go-template`.

`README.md` holds the architecture rationale and the stub-deletion checklist. `Taskfile.yml`, `.golangci.yml`, and `.github/workflows/ci.yml` are the executable source of truth — verify anything here against them.

## Project Overview

An opinionated Go **HTTP service template**: clean (DDD) architecture per feature, `samber/do/v2` DI, echo v5 HTTP, GORM over pgx, a Valkey cache-aside example, zerolog logging, and OpenTelemetry tracing + Prometheus metrics. `internal/features/example/` is a **deletable stub CRUD feature** (items) meant to be copied or removed — see README §"Deleting the stub feature".

Consumers (other Go services) import only `pkg/api/v1` to construct payloads and interpret the `{"error": code, "message": msg}` envelope; internal code never imports it.

## Architecture & Data Flow

Clean (DDD) architecture **per feature**, all dependencies pointing inward. This direction is **enforced by an executable test**, not convention alone (`internal/architecture_test.go`, build tag `unit`, runs in `task test`).

```
consumer services ──> pkg/api/v1 ──> features/*/contract    (published contract, outward-only)
handler ──> usecase.Service ──> repository.Repository <── repository/<db>
all layers ──> domain (entities + sentinel errors; standard library only)
infrastructure/* ── feature-agnostic, never imports features/**
```

A request flows: `cmd/server/main.go` loads config → `internal/app.Build` wires the do injector (`infrastructure: config → telemetry → db → valkey → server`, then each feature's `di.Register`) → route mounted at `g := e.Group("/api/v1")` → `handler` binds the contract type, calls `c.Validate`, calls `usecase.Service` → the usecase parses ids, applies business rules, maps domain↔contract → `repository.Repository` (GORM) behind an optional `CachedRepository` decorator (Valkey cache-aside) → domain sentinel errors are mapped to the HTTP envelope by `infrastructure/errors`.

**Dependency gates** — `internal/architecture_test.go` parses imports (`parser.ImportsOnly`, skipping `mock/` and `_test.go`) across 8 rules; a violation fails with `package %q violates %s`. If a change trips a rule, **restructure the change — never weaken the rule**:

| Rule | Forbids |
|---|---|
| `published-contract-is-outward-only` | `internal/**` importing `pkg/api/v1` |
| `domain-is-innermost` | anything non-stdlib in `features/<f>/domain` |
| `contract-is-leaf` | anything non-stdlib in `features/<f>/contract` |
| `repository-interface-depends-only-on-domain` | the `repository` package referencing anything but its own `domain` |
| `usecase-depends-on-domain-repository-contract` | the `usecase` package importing anything outside its own domain/repository/contract |
| `repository-impl-ignores-usecase-and-handler` | `repository/<db>/**` importing `usecase` or `handler` |
| `handler-ignores-repository` | the `handler` package importing `repository` |
| `infrastructure-ignores-features` | `infrastructure/**` importing `features/**` |

## Key Directories

- `cmd/server/` — thin entry point: loads config, sets `Version`/`CommitSHA`/`BuildTime` (ldflags), installs the signal handler, calls `app.Run`.
- `cmd/migrate/` — self-contained migration runner (`up`/`down [N]`/`force`/`version`); `fsmerge.go` unions every feature's embedded migration `fs.FS` via `migrationSources()`.
- `internal/app/` — reusable composition root (`Build`, `Run`); the only place wiring order is defined.
- `internal/architecture_test.go` — the dependency gates above.
- `internal/features/example/` — the stub feature. Layers, each its own package: `domain` (entity + sentinels), `contract` (zero-dep wire types), `usecase` (`Service` iface + `Usecase`), `repository` (outbound iface + `mock/`) and `repository/postgres/` (GORM impl, `models/`, `migrations/`, cache-aside decorator), `handler` (echo v5), `di`.
- `internal/infrastructure/` — cross-cutting: `config`, `db`, `valkey`, `errors`, `lifecycle`, `middleware`, `server`, `telemetry`.
- `internal/testutil/` — shared test helpers + `fixtures/`.
- `pkg/api/errcodes/` — published error-code constants. `pkg/api/v1/` — type-alias facade over the feature contract (published surface).
- `test/e2e/` — end-to-end tests.

## Development Commands

[Task](https://taskfile.dev) is the runner — prefer `task <name>` over raw `go` (it encodes flags you will get wrong by hand):

- `task build` / `task run` — build `bin/server` with version ldflags / run it (`run` depends on `build`).
- `task test` (alias `task test-unit`) — unit suite with coverage profile.
- `task test-integration` — needs live postgres + Valkey (`docker compose up -d postgres valkey`, then `cp .env.example .env`).
- `task test-e2e` — boots the full server (skips if the DB/Valkey TCP probe fails).
- `task lint` / `task fmt` (gofumpt + goimports) / `task tidy` / `task verify` (tidy + `git diff --exit-code go.mod go.sum`).
- `task generate` — regenerate mockgen mocks (`go generate ./...`); run after touching a repository/usecase interface.
- `task migrate-up` / `migrate-down` / `migrate-create NAME=...` — need the golang-migrate CLI on `PATH`. `go run ./cmd/migrate up` is the self-contained equivalent used in CI and containers.

**Build tags are mandatory; plain `go test ./...` runs zero tests** — every `*_test.go` is gated by `unit`, `integration`, or `e2e`:

```bash
go test -race -tags=unit -run TestName ./internal/features/example/usecase/...
go test -race -tags=integration ./internal/features/example/repository/postgres/...
go test -race -tags=unit ./internal/ -run TestArchitecture   # dependency gates only
```

The Taskfile's `dotenv` is deliberately **per-task** (`run`/`test-integration`/`test-e2e`/`migrate-*` only). Do **not** make it global: viper binds env over `config.yaml`, so a global dotenv makes `task test` assert wrong values.

## Code Conventions & Common Patterns

- **Naming.** Layer packages are fixed lowercase nouns (`domain`, `contract`, `usecase`, `repository`, `handler`, `di`). Interfaces are named by role, not feature: `usecase.Service` (inbound), `repository.Repository` (outbound). Impls: `Usecase`, `Repository`, `CachedRepository` (decorator), `Handler`, `Application`. Constructors are `New*` returning concrete types; every layer's DI entrypoint is `Register(c do.Injector) error` (or `Register(ctx, c)` when construction needs cancellation).
- **DI (samber/do/v2).** `internal/app.Build` is the composition root in fixed order; each feature wires itself via `do.Provide(...)` in `di.Register`. Providers implementing `Shutdowner*` are shut down by the container automatically — do not additionally close them in `Application.shutdown`.
- **Error handling.** Three tiers: (1) **domain sentinels** — package-level `Err<Name>` via `errors.New` in `features/*/domain/errors.go`; (2) **boundary sentinels** — `*AppError{Code, Message, HTTPStatus, Cause}` in `infrastructure/errors` (`ErrNotFound`, `ErrInvalidInput`, `ErrInternal`, …), codes from `pkg/api/errcodes` so served and published codes cannot drift; (3) **registration** — features call `apperrors.RegisterSentinel(domain.ErrX, apperrors.ErrY)` in `di`, and the handler maps any error via `status, body := apperrors.HTTPError(err); return c.JSON(status, body)`. Framework errors (404/405/413) go through the same mapper. The envelope is always `{"error": code, "message": msg}`.
- **Config.** Loaded from `config.yaml` + unprefixed env vars (viper + validator); every leaf is explicitly bound in `internal/infrastructure/config`. `CONFIG_FILE` overrides the path. `app.Build` calls `cfg.Validate()`, so bad values fail startup. `EXAMPLE_ENABLED=false` skips the stub feature's providers and routes entirely.
- **Persistence.** The SQL schema is owned by golang-migrate files under `repository/postgres/migrations/` (embedded via `//go:embed`). GORM model tags only map to existing columns — **`AutoMigrate` is never used**.
- **Generated code — regenerate, never hand-edit.** `//go:generate go tool mockgen` directives on `usecase/service.go` and `repository/repository.go` write into `*/mock/`. Tests won't compile after an interface change until `go generate ./...`. `*/mock/` is excluded from lint.
- **Echo v5 gotcha.** Handlers take `*echo.Context` (v5 changed Context from interface to struct) and return `error`. This is correct, not a typo.
- Formatting: gofumpt + goimports (`.editorconfig`: tabs for `.go`, 2-space YAML, LF, final newline). `gocyclo` max-complexity is 15; CI fails on `gofmt -s` diffs.

## Important Files

- `cmd/server/main.go` — process entry point (config, ldflags vars, signals, `app.Run`).
- `internal/app/app.go` — composition root: DI wiring order, `Build`, `Run`.
- `internal/architecture_test.go` — the 8 dependency gates (source of truth for layering).
- `internal/features/example/di/di.go` — the canonical feature wiring: flag gate, sentinel registration, providers, route mount.
- `internal/infrastructure/config/config.go` — config struct, per-leaf env binding, `Validate`.
- `internal/infrastructure/errors/` — `app_error.go`, `sentinel.go`, `mapper.go`, `status.go` (error→HTTP mapping).
- `internal/infrastructure/server/` — `http.go` (echo bootstrap, error handler) and `shutdown.go` (graceful shutdown).
- `internal/infrastructure/valkey/` — `client.go` (`NewClient`, `NewCacheAside`), `health.go`.
- `cmd/migrate/fsmerge.go` — merges per-feature embedded migrations.
- `Taskfile.yml`, `.golangci.yml`, `config.yaml`, `.env.example`, `compose.yml`, `Containerfile`, `.github/workflows/ci.yml`.

## Runtime/Tooling Preferences

- **Go 1.27** (`go.mod` directive is the effective floor; CI uses `actions/setup-go` with `go-version: stable`). Module: `github.com/zercle/zercle-go-template`. stdlib `uuid` is used directly (no third-party UUID lib).
- **Task** (taskfile.dev) is the command runner; **golang-migrate CLI** is needed only for the `task migrate-*` variants.
- Key deps: `labstack/echo/v5`, `gorm.io/gorm` + `gorm.io/driver/postgres`, `valkey-io/valkey-go` (+ `valkeyaside`), `samber/do/v2`, `spf13/viper`, `go-playground/validator/v10`, `rs/zerolog`, `golang-migrate/migrate/v4`, OTel `otel`/`sdk` (+ `otlptracehttp`, Prometheus exporter). Test stack: `stretchr/testify`, `go.uber.org/mock`, `DATA-DOG/go-sqlmock`.
- `mockgen` is declared as a Go **tool** dependency (`tool go.uber.org/mock/mockgen`) — invoked via `go generate`, not a global binary.
- Lint: **golangci-lint v2** (`.golangci.yml`) with `gocyclo` min-complexity 15, `errcheck.check-type-assertions: true`, and `wrapcheck`. Run `task fmt` and `task lint` before committing.
- Local services: `docker compose up -d postgres valkey` (`compose.yml` uses `postgres:18-alpine` and `valkey:9-alpine`). `Containerfile` is a two-stage distroless/non-root server image.

## Testing & QA

- **Frameworks.** testify (`require` for preconditions, `assert` for final checks; integration suites use `suite.Suite`); `go.uber.org/mock` generated mocks; `go-sqlmock` for GORM repository unit tests; `httptest` + `e.ServeHTTP` for handler tests. Most unit tests call `t.Parallel()`. Test names follow `Test<Type>_<Method>_<Case>`.
- **Tags.** `unit` (hermetic, mocked — 30+ files), `integration` (live postgres + Valkey, under `repository/postgres/`), `e2e` (`test/e2e/server_e2e_test.go`, boots the server, skips when infra is unreachable). **No tag = no test compiled.**
- **Architecture gate.** `go test -tags=unit ./internal/ -run TestArchitecture` — run this after any refactor that moves imports or packages.
- **Coverage.** 60% total gate, enforced in CI (`ci.yml` unit job) — not in the Taskfile. Measure locally with `task test-unit` then `go tool cover -func=coverage.out | grep total`.
- **Integration setup.** No testcontainers: tests read `config.Load()` (env-bound), apply migrations in-suite via the embedded `iofs` FS, and `TRUNCATE` between cases; they hard-fail (not skip) when infra is missing, with a production guard.
- **CI.** `.github/workflows/ci.yml`: `lint` → `unit` (60% gate, codecov, coverage artifact) → `integration` (postgres + Valkey service containers) → `build`; **lint fails on `gofmt -s` diffs and go.mod/go.sum drift**. `security.yml` runs a weekly Trivy + govulncheck scan; `cd.yml` publishes a multi-arch image on a `v*` tag.

## Gotchas

- `cp .env.example .env` (needed for integration/e2e) combined with a **global** dotenv would silently override `task test`'s asserted config values — hence per-task dotenv.
- Changing the repository/usecase interface without `task generate` leaves a stale mock and a non-compiling tree.
- The `example` feature is a stub: deleting it requires the README checklist (feature dir, `pkg/api/v1` aliases, `migrationSources()`, `app.Build`, `config.yaml`/`.env.example`, `Taskfile.yml` migrate paths).
