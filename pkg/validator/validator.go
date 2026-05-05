// Package validator wraps go-playground/validator with a process-wide instance.
package validator

import "github.com/go-playground/validator/v10"

var v = validator.New(validator.WithRequiredStructEnabled())

// Struct validates s and returns the underlying validator error if any.
func Struct(s any) error { return v.Struct(s) }

// Instance exposes the shared validator (e.g. to register custom tags).
func Instance() *validator.Validate { return v }
