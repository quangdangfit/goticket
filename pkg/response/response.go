// Package response provides a standardized JSON envelope for HTTP responses.
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/logger"
)

// Envelope is the standard response shape sent to clients.
type Envelope struct {
	Data      any    `json:"data,omitempty"`
	Error     *Error `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Error carries a stable code, human message, and optional details.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 envelope with payload.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data, RequestID: logger.RequestID(c.Request.Context())})
}

// Created writes a 201 envelope with payload.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data, RequestID: logger.RequestID(c.Request.Context())})
}

// Fail maps an error to the appropriate HTTP status and writes the envelope.
// Internal error details are not leaked to clients.
func Fail(c *gin.Context, err error) {
	status, code, msg := classify(err)
	c.JSON(status, Envelope{
		Error:     &Error{Code: code, Message: msg},
		RequestID: logger.RequestID(c.Request.Context()),
	})
}

func classify(err error) (int, string, string) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, "not_found", err.Error()
	case errors.Is(err, apperr.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized", err.Error()
	case errors.Is(err, apperr.ErrForbidden):
		return http.StatusForbidden, "forbidden", err.Error()
	case errors.Is(err, apperr.ErrConflict), errors.Is(err, apperr.ErrIdempotent):
		return http.StatusConflict, "conflict", err.Error()
	case errors.Is(err, apperr.ErrValidation), errors.Is(err, apperr.ErrInvalidPayload):
		return http.StatusBadRequest, "validation_failed", err.Error()
	case errors.Is(err, apperr.ErrRateLimited):
		return http.StatusTooManyRequests, "rate_limited", err.Error()
	case errors.Is(err, apperr.ErrSoldOut):
		return http.StatusGone, "sold_out", err.Error()
	case errors.Is(err, apperr.ErrSeatTaken):
		return http.StatusConflict, "seat_taken", err.Error()
	default:
		return http.StatusInternalServerError, "internal_error", "internal error"
	}
}
