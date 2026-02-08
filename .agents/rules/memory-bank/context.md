# Zercle Go Template - Context & Current State

## Current Work Focus

**Initial template setup and documentation completion**

The project foundation is established with Clean Architecture patterns implemented. Current priorities:

1. **Memory Bank completion**: Creating comprehensive documentation (this file set)
2. **README creation**: User-facing setup and usage guide
3. **Integration verification**: Ensuring all components work together

## Recent Changes

*No recent changes - initial project creation phase*

### Committed Decisions

| Decision | Rationale | Date |
|----------|-----------|------|
| Echo framework | Balance of performance and simplicity vs Gin/Fiber | Initial |
| sqlc over GORM | Compile-time type safety, raw SQL control | Initial |
| pgx/v5 over lib/pq | Modern driver with better performance | Initial |
| Feature-based structure | Clear boundaries, team scalability | Initial |
| Dual repository pattern | Testing without database dependency | Initial |
| Zerolog over logrus | Better performance, cleaner API | Initial |
| Viper for config | Supports multiple sources natively | Initial |

## Next Steps

### Immediate (This Week)
1. [ ] Create README.md with quickstart guide
2. [ ] Add example environment files (.env.example comprehensive)
3. [ ] Verify Docker Compose works end-to-end
4. [ ] Add integration test for auth flow

### Short-term (Next 2 Weeks)
1. [ ] Add health check endpoint
2. [ ] Implement rate limiting
3. [ ] Add request ID middleware
4. [ ] Create seed data script
5. [ ] Add pagination helpers

### Medium-term (Next Month)
1. [ ] RBAC implementation
2. [ ] Audit logging
3. [ ] Metrics endpoint (Prometheus)
4. [ ] CI/CD pipeline examples
5. [ ] Load testing setup

## Active Considerations

### Under Evaluation

| Topic | Options | Leaning Towards | Notes |
|-------|---------|-----------------|-------|
| ORM addition | sqlc only / GORM / ent | Stay with sqlc | Type safety is priority |
| Caching layer | Redis / in-memory / none | Redis for prod | Needs investigation |
| Message queue | NATS / RabbitMQ / none | NATS | Lightweight, Go-native |
| API format | REST only / add gRPC | REST for now | Complexity vs benefit |
| Config format | YAML / TOML / JSON | Keep YAML | Widely understood |

### Technical Debt Tracking

*None currently identified*

## Important Patterns to Remember

### 1. Dependency Injection Container

All dependencies are wired in [`internal/container/container.go`](internal/container/container.go). When adding new features:

```go
// 1. Add repository interface implementation choice
// 2. Update Container struct
// 3. Add functional option
// 4. Include in Build() method
```

### 2. Feature Module Structure

Every feature follows this exact structure:

```
internal/feature/{name}/
├── domain/          # Entities and business rules (zero deps)
├── dto/             # Request/response structs
├── handler/         # HTTP handlers ( Echo context)
├── repository/      # Data access interface + implementations
├── usecase/         # Business logic orchestration
└── {name}.go        # Public exports
```

### 3. Error Handling Pattern

Always use custom error types from [`internal/errors/errors.go`](internal/errors/errors.go):

```go
return nil, errors.ErrInvalidInput.WithMessage("email is required")
```

Handlers translate to HTTP status codes automatically.

### 4. Testing Strategy

- **Unit tests**: Mock dependencies with `//go:generate mockgen`
- **Integration tests**: Use `TestMain` with database setup
- **Table-driven**: All tests use table-driven pattern
- **Naming**: `TestFunctionName_Scenario` convention

### 5. Database Workflow

1. Write SQL in [`internal/infrastructure/db/queries/`](internal/infrastructure/db/queries/)
2. Update schema in [`internal/infrastructure/db/migrations/`](internal/infrastructure/db/migrations/)
3. Run `sqlc generate` (via Makefile)
4. Use generated code in repository layer

## Known Constraints

1. **Go 1.24+ required**: Uses modern Go features
2. **PostgreSQL 13+**: For JSONB and CTE support
3. **Docker required**: For consistent development environment
4. **Make required**: For build automation

## Communication Notes

- **Project owner**: Zercle organization
- **License**: Check repository for license details
- **Contributing**: Follow existing patterns, add tests for new features
- **Questions**: Refer to Memory Bank first, then code comments

## Quick Reference

| Task | Command |
|------|---------|
| Start development | `make docker-up` |
| Run tests | `make test` |
| Run linter | `make lint` |
| Generate SQLC | `make sqlc` |
| Swagger docs | `make swag` |
| Build binary | `make build` |

## Documentation Cross-References

- Architecture decisions: [`architecture.md`](architecture.md)
- Technology details: [`tech.md`](tech.md)
- Development workflows: [`tasks.md`](tasks.md)
- Project overview: [`brief.md`](brief.md)
