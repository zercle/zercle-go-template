# AGENTS.md

Repo-specific guidance for AI coding agents working in `zercle-go-template`.

`README.md` holds the architecture rationale and stub-deletion checklist; `Taskfile.yml`, `.golangci.yml`, and `.github/workflows/ci.yml` are the executable source of truth — verify anything here against them.

## Commands

[Task](https://taskfile.dev) is the runner — prefer `task <name>` over raw `go` (it encodes flags you will get wrong by hand):

- `task build` / `task run` — build `bin/server` with version ldflags / run it
- `task test` — unit suite (default); `task test-integration` (needs live postgres+valkey); `task test-e2e` (boots the server)
- `task lint` (golangci-lint v2) / `task fmt` (gofumpt + goimports) / `task tidy` + `task verify` (fails on go.mod/go.sum drift)
- `task generate` — mockgen mocks **then** protoc; run after touching a port interface or `.proto`; `task proto` for protoc alone
- `go run ./cmd/migrate up` — self-contained migration runner (CI path; the `task migrate-up` / `migrate-down` / `migrate-create NAME=...` variants need the golang-migrate CLI installed)

Single test — **build tags are mandatory; plain `go test ./...` runs zero tests:**

```bash
go test -race -tags=unit -run TestName ./internal/features/example/application/...
go test -race -tags=integration ./internal/features/example/adapter/out/postgres/...
go test -race -tags=unit ./internal/ -run TestArchitecture   # dependency gates only
```

Local services: `docker compose up -d postgres valkey`, then `cp .env.example .env` for integration/e2e. The Taskfile's `dotenv` is deliberately **per-task** (run/test-integration/test-e2e/migrate-* only) — do not make it global: viper binds env over config.yaml, so a global dotenv makes `task test` assert wrong values.

## Architecture

Clean (DDD) architecture **per feature** under `internal/features/<name>/`, cross-cutting infra under `internal/platform/`. Dependency direction — all arrows inward:

```
consumer services ──> pkg/api/v1 ──> features/*/contract    (published contract, outward-only)
adapter/in/{http,grpc} ──> application.Service ──> port.Repository <── adapter/out/postgres
all layers ──> domain (entities + sentinel errors; stdlib + uuid only)
platform/* ── feature-agnostic, never imports features/**
```

Per feature: `domain` (entities, sentinels) → `contract` (canonical inbound wire types, zero deps — the *only* source of API JSON shapes) → `application` (`Service` interface speaking contract types + `Usecase` impl; uuid parsing, business validation, domain↔contract mapping all live here) → `port` (outbound interfaces) → `adapter/in` (echo v5 HTTP binds contract directly; gRPC maps pb↔contract) → `adapter/out/postgres` (GORM repo satisfying the port structurally, plus persistence models **and SQL migrations** — migrations are feature-owned).

Things that span many files:

- **`internal/architecture_test.go`** enforces the layering by scanning imports (8 rules: facade outward-only, domain/port/contract purity, application allowlist, adapter separation, platform feature-agnosticism). If a change trips it, restructure the change — never weaken the rule.
- **Published contract**: `pkg/api/v1` is a type-alias facade over the feature's `contract` + error codes from `pkg/api/errcodes`. Internal code must never import `pkg/api/v1`. gRPC consumers use `api/pb/...` directly.
- **Wiring**: `samber/do/v2` injector; every layer exposes `Register(c *do.Injector) error`; `internal/app/app.go` is the composition root (platform registers in fixed order → then each feature's `di.Register`). Adding a feature = create the package + call its `di.Register` in `app.Build` + add its migrations FS to `migrationSources()` in `cmd/migrate/fsmerge.go` (which unions every feature's embedded migrations).
- **Error mapping**: domain sentinel errors (`features/*/domain/errors.go`) are mapped to transport via `RegisterSentinel` in each feature's `di`, resolved by `internal/platform/errors` `HTTPError`/`GRPCErr` into the `{"error": code, "message": msg}` envelope. Codes come from `pkg/api/errcodes` constants (single source, no drift).
- **Config**: `config.yaml` + unprefixed env vars, explicitly bound per leaf in `internal/platform/config`; `CONFIG_FILE` overrides the path. Schema is owned by golang-migrate SQL files, never GORM AutoMigrate.

## Generated code — regenerate, never hand-edit

- Mocks: `go:generate` directives on `application/service.go` and `port/repository.go` → `*/mock/`. Tests won't compile after touching a port until `go generate ./...`. CI runs `go generate ./...` before tests — skipping it locally means your tree diverges.
- `api/pb/**` (protoc) and `*/mock/` are excluded from lint. `task fmt` deliberately excludes `api/` (goimports regroups protoc's imports every run). Note: the committed `api/pb` predates the current local protoc version — `task generate` may show it as modified; don't commit unrelated pb churn.

## Gotchas

- **Echo v5**: handlers take `*echo.Context` (v5 changed Context from interface to struct). This is correct, not a mistake.
- CI order is lint → unit (60% coverage gate) → integration → build; lint fails on `gofmt -s` and go.mod drift. Run `task fmt` and `task lint` before committing.
- The `example` feature is a deletable stub — README §"Deleting the stub feature" lists every place it touches.
