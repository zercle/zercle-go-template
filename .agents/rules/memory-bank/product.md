# Product Context

**Last Updated:** 2026-02-10

## Problems Solved

1. **Architecture Complexity**
   - Provides a battle-tested Clean Architecture implementation
   - Eliminates architectural decision paralysis for new projects
   - Establishes clear layer boundaries and dependency rules

2. **Security Implementation**
   - OWASP-compliant password hashing (Argon2id)
   - Industry-standard JWT authentication
   - Secure configuration management with multiple sources

3. **Development Velocity**
   - Reduces boilerplate code for common patterns
   - Provides ready-to-use testing infrastructure
   - Includes mock generation for all interfaces

4. **Database Integration**
   - Type-safe SQL queries with sqlc
   - Easy switching between in-memory and PostgreSQL
   - Connection pooling and lifecycle management

## UX Goals

- **Developer Experience:** Intuitive code structure with clear conventions
- **Onboarding:** Comprehensive documentation and examples
- **Extensibility:** Easy to add new features following established patterns
- **Testing:** First-class testing support with minimal friction
- **Observability:** Structured logging for debugging and monitoring

## Key Features

### Authentication
- Email/password login with JWT tokens
- Access token (15 min default TTL) and refresh token (7 days default)
- Token validation middleware
- Optional authentication for public endpoints

### User Management
- Create, read, update, delete users
- Email uniqueness validation
- Password update with old password verification
- Paginated user listing

### Security
- Argon2id password hashing with configurable parameters
- Environment-based security defaults (production vs development)
- Constant-time password comparison
- Secure token generation with UUID jti

### API Design
- RESTful endpoints following best practices
- Standardized response format with success/error/meta
- Swagger/OpenAPI documentation at `/swagger/*`
- Request validation with detailed error messages

### Configuration
- Multi-source configuration (env vars > .env > YAML > defaults)
- Environment-aware defaults (development/production)
- Type-safe configuration with validation
- Flexible Argon2id parameter tuning

## Roadmap Considerations

### Near-term (Potential Enhancements)
- Refresh token endpoint implementation
- Token revocation/blacklist mechanism
- Rate limiting middleware
- Request context propagation improvements
- Additional authentication methods (OAuth2, etc.)

### Medium-term (Feature Expansion)
- Role-based access control (RBAC)
- User profile management
- Email verification flow
- Password reset functionality
- Audit logging for sensitive operations

### Long-term (Architecture Evolution)
- Microservice extraction patterns
- Event-driven architecture support
- Caching layer integration (Redis)
- Message queue integration
- GraphQL API option

## Acceptance Criteria

- All features must have unit tests (80%+ coverage for critical paths)
- Integration tests for database operations
- Benchmark tests for performance-critical code
- Mock generation for all interfaces using go:generate
- Swagger documentation must be up-to-date
- Code must pass golangci-lint checks
- Pre-commit hooks must validate changes
- Docker container must build and run successfully
