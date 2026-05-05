// Package http exposes the user.Service as Gin handlers.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/user"
	"github.com/quangdangfit/goticket/internal/user/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
	"github.com/quangdangfit/goticket/pkg/validator"
)

// Handler holds the user.Service dependency.
type Handler struct{ svc user.Service }

// NewHandler constructs a Handler.
func NewHandler(svc user.Service) *Handler { return &Handler{svc: svc} }

// Register handles POST /auth/register.
func (h *Handler) Register(c *gin.Context) {
	var in dto.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Register(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}

// Login handles POST /auth/login.
func (h *Handler) Login(c *gin.Context) {
	var in dto.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(c *gin.Context) {
	var in dto.RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := validator.Struct(&in); err != nil {
		response.Fail(c, apperr.ErrValidation)
		return
	}
	out, err := h.svc.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(c *gin.Context) {
	var in dto.RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), in.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}

// Me handles GET /me — requires auth middleware to have populated user_id.
func (h *Handler) Me(c *gin.Context) {
	uid, _ := c.Get("user_id")
	id, _ := uid.(string)
	if id == "" {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	p, err := h.svc.Profile(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}
