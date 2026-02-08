// Package auth provides the authentication feature implementation.
// This is the main export file for the auth feature.
//
// Import this package to access all auth feature components:
//
//	import "zercle-go-template/internal/feature/auth"
//
// Or import specific subpackages:
//
//	import "zercle-go-template/internal/feature/auth/domain"
//	import "zercle-go-template/internal/feature/auth/usecase"
//	import "zercle-go-template/internal/feature/auth/middleware"
package auth

// This file serves as documentation and a central reference point
// for the auth feature. All actual implementations are in subpackages.

//nolint:golint
import "zercle-go-template/internal/feature/auth/domain"

// JWTClaims is re-exported from domain for convenience.
type JWTClaims = domain.JWTClaims

// TokenPair is re-exported from domain for convenience.
type TokenPair = domain.TokenPair
