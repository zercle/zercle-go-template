// Package repository provides PostgreSQL implementations of task domain repositories.
//
// This package implements the Repository interface defined in the task feature
// domain layer using PostgreSQL as the backing store. It follows clean/hexagonal
// architecture principles where the domain layer defines contracts (interfaces)
// that infrastructure components implement.
//
// The main implementation is postgresRepository which:
//   - Uses SQLC-generated queries for type-safe database access
//   - Maps PostgreSQL errors to domain-specific errors (e.g., not found to ErrTaskNotFound)
//   - Converts between SQLC models and domain entities using mapper functions
//
// Usage:
//
//	import taskRepo "github.com/zercle/zercle-go-template/internal/feature/task/repository"
//
//	// In your dependency injection/container setup:
//	db, _ := pgxpool.New(ctx, connectionString)
//	repo := taskRepo.NewPostgresRepository(db)
//
// The repository expects the SQLC queries from internal/infrastructure/database/sqlc/task
// to be generated and available.
package repository
