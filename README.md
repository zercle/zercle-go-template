# zercle-go-template

Opinionated Go HTTP service template: clean (DDD) architecture, samber/do DI, OpenTelemetry tracing, Prometheus metrics, PostgreSQL via GORM, and a Valkey cache-aside example — plus a layered CRUD feature to copy or delete.

## Prerequisites

- Go 1.27+
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

The server listens on `0.0.0.0:8080`. Health probes: `/healthz`, `/readyz`, `/metrics`.

## Directory tree

```
zercle-go-template/
├── .github/
│   ├── dependabot.yml
│   └── workflows/              # ci.yml (lint/unit/integration/build), cd.yml, security.yml
├── bin/                        # build output (ignored)
├── cmd/
│   ├── migrate/main.go         # self-contained migration runner
│   └── server/main.go          # entry point: loads config, delegates to internal/app
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
│   │       │   └── out/postgres/  # GORM repository + cache aside + models + migrations
│   │       └── di/             # feature wiring
│   ├── platform/               # cross-cutting infrastructure
│   │   ├── config/             # validated viper config
│   │   ├── db/                 # gorm pool + zerolog GORM logger
│   │   ├── valkey/             # valkey client + cache-aside facade
│   │   ├── errors/             # typed errors + HTTP mapper
│   │   ├── lifecycle/          # shared resource-close adapter for DI shutdown
│   │   ├── middleware/         # recover, request-id, access-log, cors, otel
│   │   ├── server/             # echo bootstrap + graceful shutdown
│   │   └── telemetry/          # zerolog, tracer, meter, health
│   └── testutil/               # shared test helpers + fixtures
├── pkg/
│   └── api/
│       ├── errcodes/           # published error-code constants
│       └── v1/                 # published inbound contract (alias facade)
├── test/
│   └── e2e/                    # end-to-end tests (task test-e2e)
├── .editorconfig
├── .env.example
├── .gitattributes
├── .gitignore
├── .golangci.yml
├── compose.yml                 # postgres + valkey + server
├── config.yaml
├── Containerfile
├── LICENSE
├── README.md
└── Taskfile.yml
```

## Architecture overview

The template follows **clean (DDD) architecture** inside each feature, with all dependencies pointing inward:

```
consumer services ──> pkg/api/v1 ──> features/*/contract    (published contract, outward-only)
adapter/in/http ──> application.Service ──> port.Repository <── adapter/out/postgres
all layers ──> domain (entities + sentinel errors)
platform/* ── cross-cutting, never imports features/**
```

- `domain` holds entities and sentinel errors (stdlib + uuid only).
- `contract` holds the canonical inbound wire types (json/validate tags, zero dependencies) — the single source of the API shapes.
- `application` declares the inbound use-case port (`Service`, speaking contract types) and its `Usecase` implementation.
- `port` declares the outbound (driven) ports; `adapter/out/postgres` satisfies them structurally with GORM (over pgx) and owns the persistence models and SQL migrations.
- `adapter/in/http` is the driving adapter: the echo handler binds contract types directly.
- `internal/platform` consolidates cross-cutting infrastructure: config, db pool, valkey, typed errors, middleware, servers, telemetry.

**Published inbound contract.** `pkg/api/v1` is an alias facade over the feature's `contract` types plus the error codes in `pkg/api/errcodes`, so another Go service can construct payloads and interpret the `{"error": code, "message": msg}` envelope without importing server internals. Internal code never imports `pkg/api/v1`.

**Executable dependency gates.** `internal/architecture_test.go` scans imports across `internal/` and fails when a layer reaches sideways or outward: facade imports, domain/contract purity, application's allowlist, adapter separation, and platform's feature-agnosticism. It runs as part of `task test`.

Composition uses **samber/do/v2**: every layer exposes `Register(c *do.Injector) error`. `internal/app` is the reusable composition root that wires the DI container; `cmd/server/main.go` is a thin entry point that loads config, sets build-time vars (Version/CommitSHA/BuildTime), and calls `app.Run`, which bootstraps the container in dependency order:

```
platform (config → telemetry → db → valkey → server) → features
```

**Migrations are feature-owned**: each feature's SQL lives in its `adapter/out/postgres/migrations/` and is embedded per feature; `cmd/migrate` merges every feature's migrations via `migrationSources()` in `cmd/migrate/fsmerge.go`, so deleting a feature deletes its schema with it. `task migrate-up` uses the golang-migrate CLI; `go run ./cmd/migrate up` is the self-contained equivalent used in CI and containers.

Configuration is loaded from `config.yaml` and the environment (no prefix) into a typed, validated struct via spf13/viper and go-playground/validator. `EXAMPLE_ENABLED` gates the stub feature: when false its providers and routes are not registered at all. Name and page-size limits (`EXAMPLE_MAX_NAME_LENGTH`, `EXAMPLE_MAX_PAGE_SIZE`) are enforced in the application layer, so a deployment can raise them without touching request validation.

Every HTTP failure — handler errors and framework errors (404/405, body-limit 413) alike — is served in the `{"error": code, "message": msg}` envelope, with codes from `pkg/api/errcodes`.

## Caching (Valkey cache-aside)

`internal/platform/valkey` adds cache-aside reads on top of Valkey's client-side caching (`valkeyaside`): `NewCacheAside` returns a client whose `Get(ttl, key, loader)` runs the loader on a miss while concurrent misses for the same key wait, and whose entries are invalidated by the server on write. The example feature's `CachedRepository` decorates the GORM repository with it, so the application layer stays cache-unaware. `VALKEY_TTL` sets the entry lifetime.

## Deleting the stub feature

1. Remove `internal/features/example/`.
2. Replace the example aliases in `pkg/api/v1/models.go` with your feature's contract types.
3. Remove the `examplemigrations` entry from `migrationSources()` in `cmd/migrate/fsmerge.go`.
4. Remove the `exampledi.Register(injector)` call from `internal/app/app.go` (and its import of `internal/features/example/di`).
5. Delete the `example:` block from `config.yaml` and `.env.example`.
6. Update the `migrate-*` paths in `Taskfile.yml` to your feature's migrations directory.

Then add your own feature packages under `internal/features/` and wire them in `internal/app/app.go`.

## Testing

- Unit tests (hermetic, mocked): `task test` or `go test -race -tags=unit ./...`
- Integration tests (requires postgres + valkey): `task test-integration`
- End-to-end tests: `task test-e2e`
- Regenerate mocks after changing a port interface: `task generate`

## Deployment

- `Containerfile` builds a multi-stage distroless/non-root server image.
- `compose.yml` runs postgres, valkey, and server locally.
- `.github/workflows/cd.yml` publishes a multi-arch server image to ghcr.io on a `v*` tag.
- Kubernetes manifests and release automation are intentionally omitted; add them per project.
