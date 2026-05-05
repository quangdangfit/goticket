package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
)

// RegisterRoutes mounts the user routes on rg (typically /api/v1).
// auth is the JWT verifier middleware factory caller injects.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	auth := rg.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)

	me := rg.Group("/me", middleware.Auth(verifier))
	me.GET("", h.Me)
}
