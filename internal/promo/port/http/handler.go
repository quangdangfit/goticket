// Package http exposes the promo admin endpoints.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/promo"
	"github.com/quangdangfit/goticket/internal/promo/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler exposes admin POST /promos.
type Handler struct{ svc promo.Service }

// NewHandler builds a Handler.
func NewHandler(svc promo.Service) *Handler { return &Handler{svc: svc} }

// Create handles POST /admin/promos.
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreatePromoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	if err := h.svc.Create(c.Request.Context(), in); err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, gin.H{"status": "ok"})
}
