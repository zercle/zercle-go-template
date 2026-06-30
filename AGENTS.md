# AGENTS.md

Repo-specific guidance for AI coding agents working in `zercle-go-template`.
Verify anything here against `Taskfile.yml`, `.golangci.yml`, and `.github/workflows/ci.yml` — those are the executable source of truth.

## Toolchain

- Go 1.26+. Module path: `github.com/zercle/zercle-go-template`.
- [Task](https://taskfile.dev) is the command runner. Prefer `task <name>` over raw `go` commands — they encode build flags, ldflags, and coverage flags you will get wrong by hand.
- Local services (PostgreSQL 18, Valkey 9) run via containers: `docker compose up -d postgres valkey`.

## Commands

| Task | What it does |
|---|---|
| `task build` | Build `bin/server` with version ldflags (`-X main.Version/CommitSHA/BuildTime`). |
| `task run` | Build + run server. |
| `task test` / `task test-unit` | Unit tests only. **This is the default test command.** |
| `task test-integration` | Requires live postgres + valkey. |
| `task test-e2e` | Boots the full server. |
| `task lint` | `golangci-lint run --timeout=5m ./...` |
| `task fmt` | `gofumpt -w . && goimports -w .` |
| `task tidy` + `task verify` | Tidy, then fail if `go.mod`/`go.sum` have diff. |
| `task generate` | `go generate ./...` (mockgen) **then** `task proto` (protoc). |
| `task proto` | Regenerate `api/pb/` from `api/proto/`. |
| `task migrate-up` / `migrate-down` / `migrate-create NAME=...` | golang-migrate CLI against `$DB_*` env. |

Single-package / single-test (raw go, must pass the tag yourself):

```bash
go test -race -tags=unit -run TestName ./internal/features/example/service/...
go test -race -tags=integration ./internal/features/example/repository/...
```

## Test build tags are mandatory

Every test file carries `//go:build unit | integration | e2e`. **Plain `go test ./...` runs zero tests.** Always pass `-tags=unit` (or the appropriate tag). `task test` defaults to `unit`.

- `unit` — hermetic, mocked (sqlmock + `go.uber.org/mock` mocks).
- `integration` — needs postgres + valkey; migrations are run first.
- `e2e` — under `test/e2e/`, boots the whole server.

## Generated code — regenerate, never hand-edit

- Mocks: `go:generate` directives in `internal/features/example/domain/{service,repository}.go` produce `*/mock/*_mock.go` via `mockgen`. Run `go generate ./...` after touching a port interface, or tests won't compile.
- Protobuf: `api/pb/**` is generated from `api/proto/**`. Run `task proto` after editing `.proto` files.
- `api/pb/` and `*/mock/` are excluded from lint and formatting — do not "fix" them.

CI runs `go generate ./...` before tests; if you skip it locally your tree will diverge.

## CI gates (`.github/workflows/ci.yml`)

Order: **lint → unit → integration → build**. Each stage depends on the previous.

- Lint stage also runs `go mod tidy` and fails on `go.mod`/`go.sum` diff, plus `gofmt -s -l .` must be empty.
- Unit stage enforces a **60% coverage gate** on `./internal/... ./pkg/...`. Don't drop coverage.
- Integration stage runs `go run ./cmd/migrate up` (the self-contained migrator binary, not the CLI) before tests.
- Branches: `main`, `develop`, `template-v2`.

## Lint is strict (`.golangci.yml`, golangci-lint v2)

- `wrapcheck` — errors from outside the package must be wrapped (`fmt.Errorf("...: %w", err)`). Exceptions: echo `Context.JSON`, `errors.GRPCErr`, `status.Error`.
- `errcheck` with `check-type-assertions: true`.
- `exhaustive` — switch over enums must cover all cases.
- `gocyclo` ≤ 15, `funlen` ≤ 120 lines, `nestif`, `gocritic`, `gosec`, `revive`, `testifylint`.
- Run `task fmt` (gofumpt + goimports) before committing — CI checks `gofmt -s`.

## Architecture

Clean architecture per feature under `internal/features/<name>/`:

```
domain/      entities + ports (interfaces) — no external deps
repository/  GORM implementation of the outbound port (pgx driver)
service/     use-case implementation of the inbound port
handler/     http (echo v5) + grpc transport
di/          Register(injector) wires the feature into the container
dto/         request/response shapes
```

Composition root: **`internal/app/app.go`** wires the `samber/do/v2` injector in fixed order: `config → telemetry → db → valkey → shared servers → features`. Every layer exposes `Register(c *do.Injector) error`. **To add a feature, call its `di.Register(injector)` here.** `cmd/server/main.go` is a thin entry point — do not put logic in it.

Config: `config.yaml` + environment variables (no prefix) → typed struct via viper + go-playground/validator. See `.env.example` for the full key list.

## Migrations

Two equivalent paths — keep them in sync:

- `task migrate-up` — uses the `migrate` CLI pointed at `internal/infrastructure/db/migrations`.
- `go run ./cmd/migrate up` — self-contained binary that `go:embed`s the same SQL files (used by CI and the container image `Containerfile.migrate`).

Create new ones with `task migrate-create NAME=add_users` (writes to `internal/infrastructure/db/migrations`).

## The `example` feature is a stub

`internal/features/example/` + `api/proto/example/` exist only as a reference. To start a real project, delete them per `README.md` §"Deleting the stub feature" (remove the package, the proto, the `exampledi.Register` call + import in `internal/app/app.go`, and the `example:` block in `config.yaml` / `.env.example`).
