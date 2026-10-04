# zercle-go-template

Opinionated Go HTTP service template: clean (DDD) architecture, samber/do DI, OpenTelemetry tracing, Prometheus metrics, PostgreSQL via GORM, and a Valkey cache-aside example — with three layered features (catalog, machines, sales) forming a small distributed-vending-machines demo to copy or delete.

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
│   │   ├── features.go         # feature registry: the single enumeration point
│   │   ├── catalog/            # global product pool (price + stock)
│   │   ├── machines/           # vending machines + their coin banks
│   │   └── sales/              # purchases: price a sale, compose change, commit
│   │       # catalog, machines, and sales each repeat the same nine-layer
│   │       # layout below:
│   │       ├── domain/         # entities + sentinel errors (standard library only)
│   │       ├── contract/       # canonical inbound wire types
│   │       ├── usecase/        # use-case service + implementation + mocks
│   │       ├── repository/     # outbound interface + mocks
│   │       │   └── postgres/  # GORM repository + cache aside + models + migrations
│   │       ├── handler/        # echo v5 HTTP handler
│   │       └── di/             # feature wiring
│   ├── infrastructure/         # cross-cutting infrastructure
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
handler ──> usecase.Service ──> repository.Repository <── repository/postgres
all layers ──> domain (entities + sentinel errors)
infrastructure/* ── cross-cutting, never imports features/**
```

- `domain` holds entities and sentinel errors (standard library only: `uuid` is stdlib in Go 1.27+).
- `contract` holds the canonical inbound wire types (json/validate tags, zero dependencies) — the single source of the API shapes.
- `usecase` declares the inbound use-case service (`Service`, speaking contract types) and its `Usecase` implementation.
- `repository` declares the outbound (driven) interface; `repository/postgres` satisfies it structurally with GORM (over pgx) and owns the persistence models and SQL migrations.
- `handler` is the driving adapter: the echo handler binds contract types directly.
- `internal/infrastructure` consolidates cross-cutting concerns: config, db pool, valkey, typed errors, middleware, servers, telemetry.

**Published inbound contract.** `pkg/api/v1` is an alias facade over the feature's `contract` types plus the error codes in `pkg/api/errcodes`, so another Go service can construct payloads and interpret the `{"error": code, "message": msg}` envelope without importing server internals. Internal code never imports `pkg/api/v1`.

**Executable dependency gates.** `internal/architecture_test.go` scans imports across `internal/` and fails when a layer reaches sideways or outward: facade imports, domain/contract purity, usecase's allowlist, repository/handler separation, and infrastructure's feature-agnosticism. It runs as part of `task test`.

Composition uses **samber/do/v2**: every layer exposes `Register(c *do.Injector) error`. `internal/app` is the reusable composition root that wires the DI container by iterating the feature registry in `internal/features/features.go`; `cmd/server/main.go` is a thin entry point that loads config, sets build-time vars (Version/CommitSHA/BuildTime), and calls `app.Run`, which bootstraps the container in dependency order:

```
infrastructure (config → telemetry → db → valkey → server) → features
```

The registry is the **single enumeration point**: `features.List` holds one `Feature` (`Name`, `Register`, `Migrations`) per feature, `features.RegisterAll` wires the DI container, and `features.MigrationSources` feeds the migration runner. Adding or deleting a feature is therefore one entry in that list plus the feature's own directory. That order is also the migration order: `catalog` owns schema version 1, `machines` version 2, `sales` version 3.

**Migrations are feature-owned**: each feature's SQL lives in its `repository/postgres/migrations/` and is embedded per feature; `cmd/migrate` merges every registered feature's migrations via `migrationSources()` in `cmd/migrate/fsmerge.go`, which delegates to the registry, so deleting a feature deletes its schema with it. `task migrate-up` / `migrate-down` run the self-contained `go run ./cmd/migrate ...` runner; `migrate-create` still uses the golang-migrate CLI and takes the target directory from `FEATURE=<name>`. Migration version numbers are a **single namespace across all features**, not per feature: the next migration added to any feature takes the next free version.

Configuration is loaded from `config.yaml` and the environment (no prefix) into a typed, validated struct via spf13/viper and go-playground/validator. `CATALOG_ENABLED`, `MACHINES_ENABLED`, and `SALES_ENABLED` gate each feature: when false its providers and routes are not registered at all. Name/label and page-size limits (`CATALOG_MAX_NAME_LENGTH`, `CATALOG_MAX_PAGE_SIZE`, `MACHINES_MAX_LABEL_LENGTH`, `MACHINES_MAX_PAGE_SIZE`) are enforced in the usecase layer, so a deployment can raise them without touching request validation.

### Routes

| Method | Path | Feature | Purpose |
|---|---|---|---|
| POST | `/api/v1/products` | catalog | add a product to the global pool |
| GET | `/api/v1/products` | catalog | list products (paginated) |
| GET | `/api/v1/products/:id` | catalog | fetch one product |
| POST | `/api/v1/machines` | machines | register a machine with an initial coin bank |
| GET | `/api/v1/machines` | machines | list machines (paginated) |
| GET | `/api/v1/machines/:id` | machines | fetch one machine |
| POST | `/api/v1/machines/:id/bank` | machines | restock a machine's coin bank |
| POST | `/api/v1/purchases` | sales | buy a product: price, compose change, commit |

Health and observability endpoints (`/healthz`, `/readyz`, `/metrics`) are served by `internal/infrastructure/server`.

**Cross-feature boundaries.** Features never import each other; each owns its domain, contract, and repository port. `sales` consumes catalog and machines data only through its own `repository.Repository` port, whose postgres implementation reads the `catalog_products` and `machines` tables directly and commits the sale in one transaction. This is a deliberate single-database compromise — the tables are shared, but the port is the seam: a future service split replaces that one implementation without touching the sales domain or usecase. Stock is a **global pool** (decrementing a product affects every machine), while per-machine product slots are the documented extension if the demo grows.

Every HTTP failure — handler errors and framework errors (404/405, body-limit 413) alike — is served in the `{"error": code, "message": msg}` envelope, with codes from `pkg/api/errcodes`.

## Caching (Valkey cache-aside)

`internal/infrastructure/valkey` adds cache-aside reads on top of Valkey's client-side caching (`valkeyaside`): `NewCacheAside` returns a client whose `Get(ttl, key, loader)` runs the loader on a miss while concurrent misses for the same key wait, and whose entries are invalidated by the server on write. The catalog feature's `CachedRepository` decorates the GORM repository with it, so the usecase layer stays cache-unaware. `VALKEY_TTL` sets the entry lifetime.

## Adding and deleting features

To add a feature:

1. Copy or author `internal/features/<name>/` (the layers described above).
2. Add one entry to `features.List` in `internal/features/features.go` — `Name`, `Register`, and `Migrations` when it owns schema. Migrations are numbered in one global namespace, so take the next free version across all features.
3. Register its config in `internal/infrastructure/config` and add its aliases in `pkg/api/v1` before publishing.

To replace the demo features (catalog, machines, sales):

1. Remove the feature directories under `internal/features/` you are replacing.
2. Remove their entries from `features.List` in `internal/features/features.go`.
3. Replace their aliases in `pkg/api/v1/models.go` with your feature's contract types.
4. Delete the `catalog:` / `machines:` / `sales:` blocks from `config.yaml` and the matching `CATALOG_*` / `MACHINES_*` / `SALES_*` lines from `.env.example`.

## Testing

- Unit tests (hermetic, mocked): `task test` or `go test -race -tags=unit ./...`
- Integration tests (requires postgres + valkey): `task test-integration`
- End-to-end tests: `task test-e2e`
- Regenerate mocks after changing the repository interface: `task generate`

## Deployment

- `Containerfile` builds a multi-stage distroless/non-root server image.
- `compose.yml` runs postgres, valkey, and server locally.
- `.github/workflows/cd.yml` publishes a multi-arch server image to ghcr.io on a `v*` tag.
- Kubernetes manifests and release automation are intentionally omitted; add them per project.
