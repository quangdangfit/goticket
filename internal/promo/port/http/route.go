package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
	"github.com/quangdangfit/goticket/internal/user/model"
)

// RegisterRoutes mounts /admin/promos behind admin role guard.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	g := rg.Group("/admin/promos", middleware.Auth(verifier), middleware.RequireRole(model.RoleAdmin))
	g.POST("", h.Create)
}
