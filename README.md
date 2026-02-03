# 🚀 Go Template

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License"></a>
  <a href="https://github.com/zercle/zercle-go-template/actions"><img src="https://img.shields.io/badge/CI-GitHub%20Actions-blue?style=for-the-badge&logo=githubactions" alt="CI"></a>
</p>

> Production-ready Go web application template with Clean Architecture, JWT authentication, and PostgreSQL.

---

- 🏗️ **Clean Architecture** - Feature-based organization with clear separation
- 🔐 **JWT Authentication** - Secure auth with Argon2id password hashing
- 🗄️ **PostgreSQL** - Type-safe queries with sqlc
- 🌐 **Echo Framework** - High-performance minimalist web framework
- 📝 **Swagger/OpenAPI** - Auto-generated API documentation
- 🧪 **Testing** - Unit, integration, and benchmark tests
- 🐳 **Docker** - Multi-stage builds for production containers
- ⚡ **CI/CD** - Automated GitHub Actions pipeline
- 📊 **Logging** - Structured logging with Zerolog

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| Language | [Go 1.25](https://go.dev) |
| Framework | [Echo v4](https://echo.labstack.com/) |
| Database | [PostgreSQL](https://www.postgresql.org/) + [pgx v5](https://github.com/jackc/pgx) |
| Auth | [golang-jwt](https://github.com/golang-jwt/jwt) + Argon2id |
| Config | [Viper](https://github.com/spf13/viper) |
| Logging | [Zerolog](https://github.com/rs/zerolog) |
| Docs | [Swagger](https://swagger.io/) |
| Testing | [Testify](https://github.com/stretchr/testify) |

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/zercle/zercle-go-template.git
cd zercle-go-template

# Copy environment file
cp .env.example .env

# Run with Docker
docker-compose up -d

# Or run locally
make migrate-up
make run
```

API available at `http://localhost:8080`

## 📁 Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/           # Configuration (Viper)
│   ├── container/        # Dependency injection
│   ├── errors/           # Custom error types
│   ├── feature/          # Feature modules
│   │   ├── auth/         # Authentication
│   │   └── user/         # User management
│   ├── infrastructure/   # Database layer
│   ├── logger/           # Zerolog wrapper
│   └── middleware/       # HTTP middleware
├── api/docs/             # Swagger documentation
└── configs/              # Configuration files
```

## ⚙️ Configuration

Environment variables with `APP_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_SERVER_PORT` | Server port | `8080` |
| `APP_DATABASE_HOST` | Database host | `localhost` |
| `APP_DATABASE_PORT` | Database port | `5432` |
| `APP_DATABASE_DATABASE` | Database name | `zercle_template` |
| `APP_DATABASE_USERNAME` | DB username | `postgres` |
| `APP_DATABASE_PASSWORD` | DB password | `postgres` |
| `APP_JWT_SECRET` | JWT signing secret | *(required)* |

## 🧪 Testing

```bash
make test           # Run all tests
make test-coverage # Run with coverage
make benchmark     # Run benchmarks
```

## 📚 API Docs

Swagger UI: `http://localhost:8080/swagger/index.html`

### Benchmarks

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/refresh` | Refresh token |
| GET | `/api/v1/users` | List users |
| GET | `/api/v1/users/:id` | Get user |
| POST | `/api/v1/users` | Create user |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user |

## 💻 Development

```bash
make install-tools  # Install dev tools
make dev            # Run with hot reload
make lint           # Run linter
make fmt            # Format code
```

## 🐳 Docker

```bash
make docker-build   # Build image
make docker-run     # Run container
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test`
4. Submit a pull request

body (optional)

MIT License - see [LICENSE](LICENSE)
