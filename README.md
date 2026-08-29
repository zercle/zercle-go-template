# zercle-go-template

Opinionated Go microservice template with clean architecture, samber/do DI, OpenTelemetry, Prometheus metrics, and an example CRUD feature ready to be deleted.

## Prerequisites

- Go 1.26+
- Docker/Podman
- [Task](https://taskfile.dev/installation/)
- PostgreSQL 18+ (via container)
- Valkey 9+ (via container)

## Quick start

```bash
cp .env.example .env
docker compose up -d postgres valkey
task migrate-up
task run
```

The server listens on `0.0.0.0:8080` for HTTP and `0.0.0.0:50051` for gRPC.

## Directory tree

```
zercle-go-template/
├── .agents/
│   ├── AGENTS.md
│   └── plans/                  # spec/canvas/state per task
├── .github/
│   ├── dependabot.yml
│   └── workflows/
│       └── ci.yml
├── api/
│   ├── pb/example/v1/          # generated protobuf Go code (task proto)
│   └── proto/example/v1/
│       └── example.proto
├── bin/                        # build output (ignored)
├── cmd/
│   ├── migrate/main.go         # migration runner
│   └── server/main.go          # entry point: loads config, delegates to internal/app
├── deployments/
│   ├── kustomize/
│   │   ├── base/
│   │   └── overlays/
│   └── observability/          # otel-collector + prometheus configs
├── internal/
│   ├── app/                    # reusable composition root (DI wiring, app.Run)
│   ├── architecture_test.go    # executable dependency gates (runs in task test)
│   ├── features/
│   │   └── example/            # STUB FEATURE — delete to start
│   │       ├── domain/         # entities + sentinel errors (stdlib + uuid only)
│   │       ├── contract/       # canonical inbound wire types
│   │       ├── application/    # use-case port + implementation + mocks
│   │       ├── port/           # outbound ports + mocks
│   │       ├── adapter/
│   │       │   ├── in/http/    # echo v5 driving adapter
│   │       │   ├── in/grpc/    # gRPC driving adapter
│   │       │   └── out/postgres/  # GORM repository + models + migrations
│   │       └── di/             # feature wiring
│   ├── platform/               # cross-cutting infrastructure
│   │   ├── config/             # validated viper config
│   │   ├── db/                 # gorm pool
│   │   ├── valkey/             # valkey client
│   │   ├── errors/             # typed errors + mappers
│   │   ├── middleware/         # recover, request-id, access-log, cors, otel
│   │   ├── server/             # echo + grpc bootstrap, shutdown
│   │   └── telemetry/          # zerolog, tracer, meter, health
│   └── testutil/               # shared test helpers + fixtures
├── pkg/
│   ├── api/
│   │   ├── errcodes/           # published error-code constants
│   │   └── v1/                 # published inbound contract (alias facade)
│   └── uuidgen/
├── test/
│   └── e2e/                    # end-to-end tests (task test-e2e)
├── .editorconfig
├── .env.example
├── .gitattributes
├── .gitignore
├── .golangci.yml
├── .goreleaser.yml
├── compose.yml
├── config.yaml
├── Containerfile
├── Containerfile.migrate
├── LICENSE
├── README.md
└── Taskfile.yml
```

## Architecture overview

The template follows **clean (DDD) architecture** inside each feature, with all dependencies pointing inward:

```
consumer services ──> pkg/api/v1 ──> features/*/contract    (published contract, outward-only)
adapter/in/{http,grpc} ──> application.Service ──> port.Repository <── adapter/out/postgres
all layers ──> domain (entities + sentinel errors)
platform/* ── cross-cutting, never imports features/**
```

- `domain` holds entities and sentinel errors (stdlib + uuid only).
- `contract` holds the canonical inbound wire types (json/validate tags, zero dependencies) — the single source of the API shapes.
- `application` declares the inbound use-case port (`Service`, speaking contract types) and its `Usecase` implementation.
- `port` declares the outbound (driven) ports; `adapter/out/postgres` satisfies them structurally with GORM (over pgx) and owns the persistence models and SQL migrations.
- `adapter/in/{http,grpc}` are driving adapters: the echo handler binds contract types directly; the gRPC server maps protobuf ↔ contract.
- `internal/platform` consolidates cross-cutting infrastructure: config, db pool, valkey, typed errors, middleware, servers, telemetry.

**Published inbound contract.** `pkg/api/v1` is an alias facade over the feature's `contract` types plus the error codes in `pkg/api/errcodes`, so another Go service can construct payloads and interpret the `{"error": code, "message": msg}` envelope without importing server internals. Internal code never imports `pkg/api/v1`. gRPC consumers import `api/pb/...` directly.

**Executable dependency gates.** `internal/architecture_test.go` scans imports across `internal/` and fails when a layer reaches sideways or outward: facade imports, domain/contract purity, application's allowlist, adapter separation, and platform's feature-agnosticism. It runs as part of `task test`.

Composition uses **samber/do/v2**: every layer exposes `Register(c *do.Injector) error`. `internal/app` is the reusable composition root that wires the DI container; `cmd/server/main.go` is a thin entry point that loads config, sets build-time vars (Version/CommitSHA/BuildTime), and calls `app.Run`, which bootstraps the container in dependency order:

```
platform (config → telemetry → db → valkey → server) → features
```

**Migrations are feature-owned**: each feature's SQL lives in its `adapter/out/postgres/migrations/` and is embedded per feature; `cmd/migrate` merges every feature's migrations via `migrationSources()` in `cmd/migrate/fsmerge.go`, so deleting a feature deletes its schema with it.

Configuration is loaded from `config.yaml` and the environment (no prefix) into a typed, validated struct via spf13/viper and go-playground/validator.

## Deleting the stub feature

1. Remove `internal/features/example/`.
2. Remove `api/proto/example/` and `api/pb/example/`.
3. Replace the example aliases in `pkg/api/v1/models.go` with your feature's contract types.
4. Remove the `examplemigrations` entry from `migrationSources()` in `cmd/migrate/fsmerge.go`.
5. Remove the `exampledi.Register(injector)` call from `internal/app/app.go` (and its import of `internal/features/example/di`).
6. Delete the `example:` block from `config.yaml` and `.env.example`.
7. Update the `migrate-*` paths in `Taskfile.yml` to your feature's migrations directory.

Then add your own feature packages under `internal/features/` and wire them in `internal/app/app.go`.

## Testing

- Unit tests (hermetic, mocked): `task test` or `go test -race -tags=unit ./...`
- Integration tests (requires postgres + valkey): `task test-integration`
- End-to-end tests: `task test-e2e`

## Deployment

- `Containerfile` builds a multi-stage distroless/non-root server image.
- `Containerfile.migrate` builds a self-contained migration binary that embeds migrations via `go:embed`.
- `compose.yml` runs postgres, valkey, migrate, and server locally.
- Kubernetes manifests are under `deployments/kustomize/`.
- `goreleaser.yml` handles cross-platform binary releases.
