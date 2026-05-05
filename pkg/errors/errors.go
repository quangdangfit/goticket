// Package errors registers application-wide sentinel errors. Domains define
// their own sentinels next to their service code; this package holds errors
// that cross domain boundaries.
package errors

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrValidation     = errors.New("validation failed")
	ErrRateLimited    = errors.New("rate limited")
	ErrInternal       = errors.New("internal error")
	ErrSoldOut        = errors.New("sold out")
	ErrSeatTaken      = errors.New("seat taken")
	ErrIdempotent     = errors.New("idempotent replay")
	ErrInvalidPayload = errors.New("invalid payload")
)
