# zercle-go-template

Opinionated Go HTTP service template: clean (DDD) architecture, samber/do DI, OpenTelemetry tracing, Prometheus metrics, PostgreSQL via GORM, and a Valkey cache-aside example — with four layered features (catalog, machines, sales, reporting) forming a small distributed-vending-machines demo to copy or delete.

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
│   ├── migrate/                # self-contained migration runner (embeds migrations/)
│   └── server/                 # entry point: loads config, delegates to internal/app
├── internal/
│   ├── app/                    # THE composition root: Build/Run + one wiring file per feature + the feature list
│   │   ├── features.go         # the ordered feature list
│   │   └── <f>.go              # per-feature wiring (catalog.go, machines.go, sales.go, reporting.go)
│   ├── archguard/              # executable dependency gates (runs in task test)
│   │   ├── layering_test.go    # dependency direction between the layers
│   │   └── imports_test.go     # what a package may import at all
│   ├── features/
│   │   ├── catalog/            # global product pool (price + stock)
│   │   ├── machines/           # vending machines + their coin banks
│   │   ├── sales/              # purchases: price a sale, compose change, commit
│   │   └── reporting/          # read-only cross-feature GET /api/v1/reports/summary (no migrations)
│   │       # catalog, machines, sales, and reporting each repeat the same
│   │       # layer layout below (reporting owns no schema):
│   │       ├── domain/         # entities + sentinel errors (standard library only)
│   │       ├── port/           # outbound interface + port/mock/ generated doubles
│   │       ├── usecase/        # business logic, one operation per file + usecase/mock/
│   │       └── handler/        # driving adapter: binds the wire contract, maps the response
│   ├── infrastructure/         # cross-cutting infrastructure
│   │   ├── config/             # validated viper config
│   │   ├── postgres/           # the postgres client (connection, logger, health, shutdowner)
│   │   │   └── <f>/           # the driven adapter of feature <f> (GORM repository + models/)
│   │   ├── valkey/             # valkey client + cache-aside facade
│   │   ├── lifecycle/          # shared resource-close adapter for DI shutdown
│   │   ├── telemetry/          # zerolog, tracer, meter, health
│   │   └── transport/          # router + server lifecycle
│   │       └── middleware/     # recover, request-id, access-log, cors, otel
│   ├── shared/
│   │   └── errors/             # typed boundary errors + HTTP mapper (feature-agnostic)
│   └── testutil/               # shared test helpers + fixtures
├── pkg/
│   └── api/
│       ├── errcodes/           # published error-code constants
│       └── v1/                 # the published wire contract (canonical request/response types)
├── migrations/                 # versioned SQL (one directory for the whole service)
├── test/
│   └── e2e/                    # end-to-end tests (task test-e2e)
├── deployments/
│   └── observability/          # otel-collector-config.yaml, prometheus.yml
├── .editorconfig
├── .env.example
├── .gitattributes
├── .gitignore
├── .golangci.yml
├── compose.yml                 # postgres + valkey + migrate + server
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
consumer services ──> pkg/api/v1    (published wire contract)
handler ──> usecase.Service ──> port.Repository <── infrastructure/postgres/<f>
all layers ──> domain (entities + sentinel errors)
infrastructure/* and shared/* ── feature-agnostic, never import features/**
```

- `domain` holds entities and sentinel errors (standard library only: `uuid` is stdlib in Go 1.27+).
- `port` declares the outbound (driven) interface the feature calls, with its generated doubles in `port/mock/`.
- `usecase` speaks **domain types only** — it takes and returns domain values, never wire types — with one file per operation (`purchase.go`, `create_product.go`, …) and its generated doubles in `usecase/mock/`.
- `handler` is the driving adapter: the echo handler binds and validates the `pkg/api/v1` request, translates it into the usecase's domain inputs, and maps the domain result back to the `pkg/api/v1` response.
- `internal/infrastructure/postgres/<f>` is the driven adapter of feature `<f>`: the GORM repository, its persistence models, and the cache-aside decorator. `internal/infrastructure/postgres` also holds the postgres client itself.
- `internal/infrastructure` consolidates cross-cutting technical plumbing: config, postgres, valkey, lifecycle, telemetry, and transport (router + server lifecycle, with `transport/middleware/`).
- `internal/shared` holds code every feature uses: the typed boundary errors and the HTTP mapper that turns them into the response envelope.

**Published wire contract.** `pkg/api/v1` is the canonical source of the HTTP JSON shapes, plus the error codes in `pkg/api/errcodes`, so another Go service can construct payloads and interpret the `{"error": code, "message": msg}` envelope without importing server internals. It stays a leaf package; internal code imports it to speak the wire shapes.

**Executable dependency gates.** `internal/archguard/` scans imports across `internal/` and `pkg/api/` and fails when a layer reaches sideways or outward: domain purity, usecase/port separation, handler separation, the driven adapters' independence, infrastructure's feature-agnosticism, shared's feature-agnosticism, and the published contract's leaf-ness. `layering_test.go` holds the direction rules, `imports_test.go` the import-surface rules. It runs as part of `task test`.

Composition uses **samber/do/v2**. `internal/app` is the **single composition root**: it holds `Build`/`Run`, one wiring file per feature (`catalog.go`, `machines.go`, `sales.go`, `reporting.go`), and the ordered feature list in `features.go`. `cmd/server/main.go` is a thin entry point that loads config, sets build-time vars (Version/CommitSHA/BuildTime), and calls `app.Run`, which bootstraps the container in dependency order:

```
infrastructure (config → telemetry → postgres → valkey → transport) → features
```

The feature list in `internal/app/features.go` is the **single enumeration point**: one entry per feature in registration order. Adding or deleting a feature is therefore one entry in that list plus the feature's own directory and its wiring file.

**Migrations are service-owned**: all SQL lives in the root `migrations/` directory as one versioned namespace, embedded by `cmd/migrate` via `go:embed`. `task migrate-up` / `migrate-down` run the self-contained `go run ./cmd/migrate ...` runner; `task migrate-create NAME=<name>` writes into `migrations/` via the golang-migrate CLI. Migration version numbers are a **single namespace across the whole service**: the next migration added takes the next free version.

Configuration is loaded from `config.yaml` and the environment (no prefix) into a typed, validated struct via spf13/viper and go-playground/validator. `CATALOG_ENABLED`, `MACHINES_ENABLED`, `SALES_ENABLED`, and `REPORTING_ENABLED` gate each feature: when false its providers and routes are not registered at all. Name/label and page-size limits (`CATALOG_MAX_NAME_LENGTH`, `CATALOG_MAX_PAGE_SIZE`, `MACHINES_MAX_LABEL_LENGTH`, `MACHINES_MAX_PAGE_SIZE`) are enforced in the usecase layer, so a deployment can raise them without touching request validation. The reporting top-machines bounds (`REPORTING_DEFAULT_TOP_MACHINES`, `REPORTING_MAX_TOP_MACHINES`) are enforced the same way.

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
| GET | `/api/v1/reports/summary` | reporting | cross-feature totals + top machines by revenue |

Health and observability endpoints (`/healthz`, `/readyz`, `/metrics`) are served by `internal/infrastructure/transport`.

**Cross-feature boundaries.** Features never import each other; each owns its domain, port, usecase, and handler. `sales` consumes catalog and machines data only through its own `port.Repository` port, whose postgres implementation in `internal/infrastructure/postgres/sales` reads the `catalog_products` and `machines` tables directly and commits the sale in one transaction. This is a deliberate single-database compromise — the tables are shared, but the port is the seam: a future service split replaces that one implementation without touching the sales domain or usecase. Stock is a **global pool** (decrementing a product affects every machine), while per-machine product slots are the documented extension if the demo grows. `reporting` demonstrates the read-only side of the same seam: it aggregates totals across all three features' tables through its own port and owns no schema of its own.

Every HTTP failure — handler errors and framework errors (404/405, body-limit 413) alike — is served in the `{"error": code, "message": msg}` envelope, with codes from `pkg/api/errcodes`.

## Caching (Valkey cache-aside)

`internal/infrastructure/valkey` adds cache-aside reads on top of Valkey's client-side caching (`valkeyaside`): `NewCacheAside` returns a client whose `Get(ttl, key, loader)` runs the loader on a miss while concurrent misses for the same key wait, and whose entries are invalidated by the server on write. The catalog feature's `CachedRepository` decorates the GORM repository with it, so the usecase layer stays cache-unaware. `VALKEY_TTL` sets the entry lifetime.

## Adding and deleting features

To add a feature:

1. Copy or author `internal/features/<name>/` (the layers described above: `domain`, `port`, `usecase`, `handler`).
2. Add one entry to the feature list in `internal/app/features.go`.
3. Add its wiring file `internal/app/<name>.go` (flag gate, sentinel registration, providers, route mount).
4. If it owns schema: add SQL files in the root `migrations/` directory (versioned in one global namespace), its GORM adapter under `internal/infrastructure/postgres/<name>/`, and its wire types in `pkg/api/v1`.
5. Register its config in `internal/infrastructure/config`.

To replace the demo features (catalog, machines, sales, reporting):

1. Remove the feature directories under `internal/features/` you are replacing.
2. Remove their entries from the feature list in `internal/app/features.go` and their wiring files from `internal/app/`.
3. Replace their types in `pkg/api/v1` with your feature's wire types.
4. Delete their SQL files from `migrations/`, their adapters from `internal/infrastructure/postgres/<name>/`, and the `catalog:` / `machines:` / `sales:` / `reporting:` blocks from `config.yaml` plus the matching `CATALOG_*` / `MACHINES_*` / `SALES_*` / `REPORTING_*` lines from `.env.example`.

## Testing

- Unit tests (hermetic, mocked): `task test` or `go test -race -tags=unit ./...`
- Integration tests (requires postgres + valkey): `task test-integration`
- End-to-end tests: `task test-e2e`
- Regenerate mocks after changing a port or usecase interface: `task generate`

## Deployment

- `Containerfile` builds a multi-stage distroless/non-root server image.
- `compose.yml` runs postgres, valkey, migrate, and server locally; the `observability` profile adds OTel Collector, Prometheus, and Grafana (`docker compose --profile observability up`).
- `Containerfile.migrate` builds the migration image; the `migrate` service runs before `server` starts.
- Observability config lives under `deployments/observability/` (`otel-collector-config.yaml`, `prometheus.yml`).
- `.github/workflows/cd.yml` publishes a multi-arch server image to ghcr.io on a `v*` tag.
- Kubernetes manifests and release automation are intentionally omitted; add them per project.
