// Package http exposes the ticket.Service over Gin.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/ticket"
	"github.com/quangdangfit/goticket/internal/ticket/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler wraps the ticket.Service.
type Handler struct{ svc ticket.Service }

// NewHandler constructs a Handler.
func NewHandler(svc ticket.Service) *Handler { return &Handler{svc: svc} }

// ListByShowtime handles GET /showtimes/:id/tickets.
func (h *Handler) ListByShowtime(c *gin.Context) {
	rows, err := h.svc.ListByShowtime(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"items": rows})
}

// Create handles POST /admin/showtimes/:id/tickets.
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreateTicketTypeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Create(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}
