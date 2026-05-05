// Package http exposes the order.Service over Gin.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/order"
	"github.com/quangdangfit/goticket/internal/order/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler holds order.Service.
type Handler struct{ svc order.Service }

// NewHandler constructs a Handler.
func NewHandler(svc order.Service) *Handler { return &Handler{svc: svc} }

// Checkout handles POST /orders.
func (h *Handler) Checkout(c *gin.Context) {
	var in dto.CheckoutInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	uid, _ := c.Get("user_id")
	out, err := h.svc.Checkout(c.Request.Context(), str(uid), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}

// Get handles GET /orders/:id.
func (h *Handler) Get(c *gin.Context) {
	uid, _ := c.Get("user_id")
	out, err := h.svc.Get(c.Request.Context(), str(uid), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

// Cancel handles POST /orders/:id/cancel.
func (h *Handler) Cancel(c *gin.Context) {
	uid, _ := c.Get("user_id")
	if err := h.svc.Cancel(c.Request.Context(), str(uid), c.Param("id")); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}

func str(v any) string { s, _ := v.(string); return s }
