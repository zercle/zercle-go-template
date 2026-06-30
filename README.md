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
│   ├── config/                 # validated viper config
│   ├── features/
│   │   └── example/            # STUB FEATURE — delete to start
│   ├── infrastructure/
│   │   ├── db/                 # gorm db, migrations
│   │   └── messaging/          # valkey client
│   ├── shared/
│   │   ├── errors/             # typed errors + mappers
│   │   ├── middleware/         # recover, request-id, access-log, cors, otel
│   │   ├── server/             # echo + grpc bootstrap, shutdown
│   │   └── telemetry/          # zerolog, tracer, meter, health
│   └── testutil/               # shared test helpers + fixtures
├── pkg/
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

The template follows **clean architecture** inside each feature: `domain` defines entities and ports, `repository` implements the outbound port with GORM (over pgx), `service` implements the inbound use-case port, and `handler` exposes HTTP (echo) and gRPC endpoints.

Composition uses **samber/do/v2**: every layer exposes `Register(c *do.Injector) error`. `internal/app` is the reusable composition root that wires the DI container; `cmd/server/main.go` is a thin entry point that loads config, sets build-time vars (Version/CommitSHA/BuildTime), and calls `app.Run`, which bootstraps the container in dependency order:

```
config → telemetry → infrastructure (db, valkey) → shared servers → features
```

Configuration is loaded from `config.yaml` and the environment (no prefix) into a typed, validated struct via spf13/viper and go-playground/validator.

## Deleting the stub feature

1. Remove `internal/features/example/`.
2. Remove `api/proto/example/`.
3. Remove the `exampledi.Register(injector)` call from `internal/app/app.go` (and its import of `internal/features/example/di`).
4. Delete the `example:` block from `config.yaml` and `.env.example`.

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
