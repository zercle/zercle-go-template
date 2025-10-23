# Active Context

## Current State
- Project is a "template" state, ready for cloning/instantiation.
- **Recent Changes**:
    - Migrated logging from `slog` to `zerolog`.
    - Migrated password hashing from `bcrypt` to `argon2id`.
    - Integrated `samber/oops` for error handling.
    - Updated `golangci-lint` configuration and versions.
    - Verified project integrity (tests & linting passing).
    - **Note**: `go test -race` skipped on Windows due to missing logic (requires GCC).

## Open Questions/Issues
- **Readme Discrepancy**: `README.md` still mentions `slog` despite the migration to `zerolog`. Needs update.

## Roadmap
- Maintain template freshness (Go versions, dependency updates).
- Ensure all CI workflows remain green.
