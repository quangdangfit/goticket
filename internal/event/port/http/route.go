package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
	"github.com/quangdangfit/goticket/internal/user/model"
)

// RegisterRoutes mounts public read routes and admin write routes.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	rg.GET("/events", h.List)
	rg.GET("/events/:id", h.Detail)

	admin := rg.Group("/admin/events", middleware.Auth(verifier), middleware.RequireRole(model.RoleAdmin))
	admin.POST("", h.Create)
	admin.PATCH("/:id", h.Update)
}
