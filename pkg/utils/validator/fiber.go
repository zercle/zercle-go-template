// Package validator provides Fiber-compatible validation adapter.
package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// FiberValidator is a Fiber-compatible validator wrapper.
type FiberValidator struct {
	validator *validator.Validate
}

// NewFiberValidator creates a new Fiber validator instance.
func NewFiberValidator() *FiberValidator {
	return &FiberValidator{
		validator: GetValidator(),
	}
}

// Validate validates the struct and returns a Fiber error if validation fails.
func (v *FiberValidator) Validate(data interface{}) error {
	if err := v.validator.Struct(data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, formatValidationError(err).Error())
	}
	return nil
}

// ValidateStruct is a convenience function to validate any struct.
func ValidateStruct(data interface{}) error {
	return Validate(data)
}
