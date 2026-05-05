// Package http exposes a thin Hold/Release surface for cart-style flows.
// In normal checkout the order service calls Inventory directly; this
// surface is used when the client wants an explicit pre-checkout cart.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/inventory"
	"github.com/quangdangfit/goticket/internal/inventory/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler exposes Hold/Release endpoints.
type Handler struct{ inv inventory.Inventory }

// NewHandler builds a Handler.
func NewHandler(inv inventory.Inventory) *Handler { return &Handler{inv: inv} }

// CreateHold handles POST /holds.
func (h *Handler) CreateHold(c *gin.Context) {
	var in dto.HoldInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	uid, _ := c.Get("user_id")
	in.UserID, _ = uid.(string)
	out, err := h.inv.Hold(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}

// DeleteHold handles DELETE /holds/:id.
func (h *Handler) DeleteHold(c *gin.Context) {
	if err := h.inv.Release(c.Request.Context(), c.Param("id")); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}
