// Package http exposes the event.Service over Gin.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/event"
	"github.com/quangdangfit/goticket/internal/event/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler holds the event.Service.
type Handler struct{ svc event.Service }

// NewHandler constructs a Handler.
func NewHandler(svc event.Service) *Handler { return &Handler{svc: svc} }

// List handles GET /events.
func (h *Handler) List(c *gin.Context) {
	var q dto.EventQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"items": rows, "total": total})
}

// Detail handles GET /events/:id.
func (h *Handler) Detail(c *gin.Context) {
	e, err := h.svc.Detail(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, e)
}

// Create handles POST /admin/events (admin-only).
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreateEventInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}

// Update handles PATCH /admin/events/:id.
func (h *Handler) Update(c *gin.Context) {
	var in dto.UpdateEventInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Update(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}
