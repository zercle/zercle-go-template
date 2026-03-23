// Package repository provides PostgreSQL implementations of user domain repositories.
//
// This package implements the Repository interface defined in the user feature
// domain layer using PostgreSQL as the backing store. It follows clean/hexagonal
// architecture principles where the domain layer defines contracts (interfaces)
// that infrastructure components implement.
//
// The main implementation is postgresRepository which:
//   - Uses SQLC-generated queries for type-safe database access
//   - Maps PostgreSQL errors to domain-specific errors (e.g., unique violations
//     to ErrDuplicateEmail, not found to ErrUserNotFound)
//   - Converts between SQLC models and domain entities using mapper functions
//
// Usage:
//
//	import userRepo "github.com/zercle/zercle-go-template/internal/feature/user/repository"
//
//	// In your dependency injection/container setup:
//	db, _ := pgxpool.New(ctx, connectionString)
//	repo := userRepo.NewPostgresRepository(db)
//
// The repository expects the SQLC queries from internal/infrastructure/database/sqlc/user
// to be generated and available.
package repository
