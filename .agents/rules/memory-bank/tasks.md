# Common Workflows & Tasks

## Development
- **Start Dev Server**: `make run` (or `air` if configured).
- **Generate Code**: `make generate` (Runs sqlc, mocks, swagger).
- **Database Migrations**: `make migrate-up`, `make migrate-down`.

## Verification
- **Run Tests**: `make test` or `go test -v ./...`.
- **Lint**: `golangci-lint run --fix ./...`.

## CI/CD
- **Build Docker**: `docker build .`
- **Check Integrity**:
    1. `go generate ./...`
    2. `golangci-lint run --fix ./...`
    3. `go clean -testcache`
    4. `go test -v -race ./...`
