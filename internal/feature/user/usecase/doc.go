// Package usecase implements the business logic layer for the user feature.
// It follows the clean/hexagonal architecture pattern where this package
// contains the application-specific business rules and use cases.
//
// The usecase layer depends on the domain layer (entity, repository interfaces)
// and implements the Usecase interface defined at the feature root level.
//
// Key responsibilities:
//   - Application-specific business logic
//   - Input validation and transformation
//   - Coordination between domain entities and repository
//   - Mapping between domain entities and DTOs
//
// This package is internally used by the handler layer and is not exported
// outside the module. External packages interact with user features through
// the user.Usecase interface.
package usecase
