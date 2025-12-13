# Product Requirements - Zercle Go Fiber Template

## Vision Statement
A production-ready, high-performance Go microservice template that implements industry best practices for building scalable, maintainable web applications using Clean Architecture and Domain-Driven Design principles.

## Goals

### Primary Goals
1. **Developer Productivity**: Fast project setup with battle-tested architecture
2. **Production Ready**: Enterprise-grade patterns and security from day one
3. **Performance**: Ultra-fast HTTP handling with Fiber framework
4. **Maintainability**: Clear separation of concerns with Clean Architecture
5. **Scalability**: Designed for horizontal and vertical scaling
6. **Security**: JWT authentication, validation, and security middleware

### Non-Goals
- Monolithic UI framework
- Specific business domain (intentionally generic template)
- Multi-language support (Go only)
- Cloud-specific deployment (cloud-agnostic)

## Target Users

### Primary Users
1. **Backend Developers**: Building Go microservices
2. **Tech Leads**: Establishing architecture standards
3. **DevOps Engineers**: Deploying and maintaining services
4. **CTOs**: Setting up development teams

### User Personas
- **Fast-Paced Startup**: Needs production-ready code quickly
- **Enterprise Team**: Requires best practices and maintainability
- **Freelancer**: Wants proven patterns without learning curve
- **Learning Developer**: Wants to learn Clean Architecture

## Core Features

### 1. HTTP API Framework
- **Fiber Framework**: High-performance HTTP server
- **RESTful Endpoints**: Standard HTTP methods and status codes
- **JSON Responses**: JSend format for consistency
- **Swagger Documentation**: Auto-generated OpenAPI docs
- **CORS Support**: Configurable cross-origin policies

### 2. Authentication & Authorization
- **JWT Tokens**: Stateless authentication
- **Registration**: New user creation
- **Login**: User authentication
- **Password Security**: bcrypt hashing
- **Protected Routes**: Middleware-based auth

### 3. Database Integration
- **PostgreSQL**: Primary database
- **UUIDv7**: Time-ordered IDs for scalability
- **sqlc**: Type-safe SQL queries
- **Migrations**: Versioned database schema changes
- **Connection Pooling**: Configurable connection limits

### 4. Domain Model
- **User Entity**: Core user with ID, name, email, password
- **Post Entity**: Example entity with CRUD operations
- **Repository Pattern**: Abstracted data access
- **Service Layer**: Business logic orchestration

### 5. Configuration Management
- **Environment-Based**: Dev, UAT, Production configs
- **YAML Config Files**: Structured configuration
- **Environment Variables**: Override capability
- **Validation**: Configuration validation on startup
- **Security**: Production checks (e.g., JWT secret)

### 6. Observability
- **Structured Logging**: Contextual log entries
- **Request ID**: Correlation across logs
- **Health Checks**: Readiness and liveness probes
- **Metrics Ready**: OpenTelemetry preparation
- **Multiple Log Levels**: Debug, info, warn, error

### 7. Development Tools
- **Makefile**: Common tasks automation
- **Docker Support**: Containerization
- **Code Generation**: sqlc, mocks, swagger
- **Testing Suite**: Unit, integration, coverage
- **Linting**: golangci-lint integration

### 8. Modular Build System
- **Build Tags**: Conditional compilation for modular deployments
- **Route Modularization**: Separate route files for better organization
- **DI Hooks**: Modular dependency injection registration
- **Selective Deployment**: Build minimal binaries with only required handlers
- **Reduced Binary Size**: Smaller deployments with targeted functionality

## Functional Requirements

### User Management
- **Registration**: Create new user account
  - Input: name, email, password
  - Validation: email format, password strength
  - Output: user (without password), JWT token
- **Login**: Authenticate existing user
  - Input: email, password
  - Output: JWT token, user info
- **Profile**: Get current user profile
  - Authentication: Required
  - Output: user details (name, email, createdAt)

### Post Management
- **Create Post**: Authenticated user creates post
  - Input: title, content
  - Output: created post with ID
- **List Posts**: Public endpoint to list all posts
  - Output: paginated list of posts
- **Get Post**: Public endpoint to get post by ID
  - Input: post ID
  - Output: post details

### Health Monitoring
- **Health Check**: Readiness probe
  - Endpoint: GET /health
  - Checks: Database connectivity
- **Liveness Check**: Container health
  - Endpoint: GET /health/live
  - Checks: Application running

## Non-Functional Requirements

### Performance
- **Response Time**: <100ms for simple queries
- **Throughput**: 10,000+ requests/second (with proper hardware)
- **Database**: Connection pooling with configurable limits
- **Memory**: Efficient allocation, minimal GC pressure

### Scalability
- **Horizontal Scaling**: Stateless design
- **Database**: Connection pooling, prepared statements
- **IDs**: UUIDv7 for distributed systems
- **Caching**: Ready for Redis integration

### Security
- **Authentication**: JWT with configurable expiration
- **Password Hashing**: bcrypt with cost factor
- **Input Validation**: All inputs validated
- **SQL Injection**: Parameterized queries only
- **CORS**: Configurable origin policies
- **Rate Limiting**: Prevent abuse
- **Secrets**: Environment variables for sensitive data

### Reliability
- **Error Handling**: Explicit, no silent failures
- **Transactions**: ACID compliance for data integrity
- **Migrations**: Versioned, reversible database changes
- **Health Checks**: Application and database monitoring

### Maintainability
- **Clean Architecture**: Layer separation
- **Dependency Injection**: Loose coupling
- **Testing**: Unit, integration, coverage
- **Documentation**: Code comments, swagger, README
- **Code Generation**: sqlc, mocks
- **Linting**: Static analysis

### Developer Experience
- **Quick Start**: <5 minutes to running
- **Makefile**: Common tasks simplified
- **Docker**: One-command infrastructure
- **Hot Reload**: Development mode
- **API Docs**: Auto-generated Swagger UI
- **Error Messages**: Clear, actionable

## API Specifications

### Response Format (JSend)
```json
{
  "status": "success|error|fail",
  "data": {},
  "message": "Optional message"
}
```

### Status Codes
- 200: Success
- 201: Created
- 400: Bad Request (validation error)
- 401: Unauthorized (missing/invalid token)
- 403: Forbidden (insufficient permissions)
- 404: Not Found
- 409: Conflict (duplicate email)
- 422: Unprocessable Entity (validation failed)
- 500: Internal Server Error

### Authentication
- **Type**: Bearer token in Authorization header
- **Format**: `Authorization: Bearer <token>`
- **Expiration**: Configurable (default 24h)
- **Secret**: Environment variable

## Database Schema

### Users Table
- id: UUIDv7 (primary key)
- name: VARCHAR(255)
- email: VARCHAR(255) (unique, indexed)
- password: VARCHAR(255) (hashed)
- created_at: TIMESTAMP
- updated_at: TIMESTAMP

### Posts Table
- id: UUIDv7 (primary key)
- title: VARCHAR(255)
- content: TEXT
- user_id: UUIDv7 (foreign key to users)
- created_at: TIMESTAMP
- updated_at: TIMESTAMP

## Environment Configurations

### Development
- **Port**: 3000
- **Environment**: dev
- **Log Level**: debug
- **Database**: localhost:5432
- **CORS**: All origins (development only)

### UAT
- **Port**: Configurable
- **Environment**: test
- **Log Level**: info
- **Database**: UAT database
- **CORS**: Restricted origins

### Production
- **Port**: Configurable
- **Environment**: production
- **Log Level**: warn
- **Database**: Production cluster
- **CORS**: Strict origins
- **JWT Secret**: Must be changed from default

## Constraints

### Technical Constraints
- **Go Version**: 1.25.0 minimum
- **PostgreSQL**: 18+ for UUIDv7 support
- **Memory**: Minimal 512MB
- **CPU**: Minimal 1 core

### Architectural Constraints
- **Clean Architecture**: Strictly enforced
- **No ORM**: Use sqlc with raw SQL
- **No Global State**: Except configuration
- **Dependency Injection**: Required for dependencies

### Security Constraints
- **No Plaintext Passwords**: Must hash
- **No Hardcoded Secrets**: Environment variables only
- **HTTPS**: Required in production
- **Input Validation**: All user inputs

## Acceptance Criteria

### User Registration
- [ ] Email validation works
- [ ] Password hashing with bcrypt
- [ ] Duplicate email returns 409
- [ ] Returns JWT token
- [ ] Password not in response

### User Login
- [ ] Valid credentials return JWT
- [ ] Invalid credentials return 401
- [ ] Token expiration respected
- [ ] Returns user profile

### Protected Endpoints
- [ ] Require valid JWT
- [ ] Return 401 for missing token
- [ ] Return 401 for invalid token
- [ ] Allow requests with valid token

### Database Operations
- [ ] Transactions maintain ACID
- [ ] Migrations run successfully
- [ ] Connection pool configured
- [ ] Queries are parameterized

### Performance
- [ ] <100ms response for simple queries
- [ ] No memory leaks under load
- [ ] Database connection pool efficient
- [ ] Proper timeouts configured

## Future Roadmap

### Phase 1 (Current)
- [x] Clean Architecture foundation
- [x] User authentication
- [x] Post CRUD
- [x] PostgreSQL integration
- [x] Docker support
- [x] Swagger documentation
- [x] Build Tags System
- [x] Route Modularization
- [x] DI Hooks System
- [x] Memory Bank Documentation

### Phase 2 (Next)
- [ ] Redis caching layer
- [ ] Background jobs (Redis Queue)
- [ ] Email notifications
- [ ] File upload support
- [ ] Search functionality

### Phase 3 (Future)
- [ ] GraphQL API
- [ ] WebSocket support
- [ ] gRPC integration
- [ ] Multi-tenancy
- [ ] Audit logging
- [ ] Data partitioning

### Phase 4 (Long-term)
- [ ] Event sourcing
- [ ] CQRS implementation
- [ ] Service mesh ready
- [ ] Kubernetes manifests
- [ ] Helm charts
- [ ] Observability stack (Prometheus, Grafana)

## Success Metrics

### Development Velocity
- Time to first successful deployment: <1 day
- New developer onboarding: <2 hours
- Feature implementation time: Baseline established

### Code Quality
- Test coverage: >80%
- Linting issues: 0
- Technical debt: Minimal, tracked

### Performance
- P95 latency: <200ms
- Uptime: 99.9%
- Error rate: <0.1%

### Maintenance
- Security updates: Monthly
- Dependency updates: Monthly
- Documentation freshness: Always current

## Risk Assessment

### Technical Risks
- **Fiber Framework Maturity**: Mitigation - Express.js proven patterns
- **UUIDv7 Adoption**: Mitigation - Simple migration path
- **No ORM Complexity**: Mitigation - sqlc provides type safety

### Operational Risks
- **Configuration Complexity**: Mitigation - Strong validation
- **Database Dependencies**: Mitigation - Clear migration guides
- **Security**: Mitigation - Industry-standard practices

### Mitigation Strategies
- Comprehensive testing suite
- Clear documentation
- Strong type system
- Explicit error handling
- Security best practices
