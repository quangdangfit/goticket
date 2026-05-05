package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
	"github.com/quangdangfit/goticket/internal/user/model"
)

// RegisterRoutes mounts public + admin ticket routes.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	rg.GET("/showtimes/:id/tickets", h.ListByShowtime)
	admin := rg.Group("/admin/showtimes/:id/tickets",
		middleware.Auth(verifier), middleware.RequireRole(model.RoleAdmin))
	admin.POST("", h.Create)
}
