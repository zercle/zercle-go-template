// Package user provides the user feature implementation.
// This is the main export file for the user feature.
//
// Import this package to access all user feature components:
//
//	import "zercle-go-template/internal/feature/user"
//
// Or import specific subpackages:
//
//	import "zercle-go-template/internal/feature/user/domain"
//	import "zercle-go-template/internal/feature/user/usecase"
//	import "zercle-go-template/internal/feature/user/handler"
//	import "zercle-go-template/internal/feature/user/dto"
//	import "zercle-go-template/internal/feature/user/repository"
package user

import "zercle-go-template/internal/feature/user/domain"

// This file serves as documentation and a central reference point
// for the user feature. All actual implementations are in subpackages.

// User is re-exported from domain for convenience.
type User = domain.User

// DomainError is re-exported from domain for convenience.
type DomainError = domain.DomainError
