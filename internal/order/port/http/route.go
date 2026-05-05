package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
)

// RegisterRoutes mounts /orders endpoints (auth required). rl is an
// optional rate-limit handler (e.g. token bucket per user).
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier, rl gin.HandlerFunc) {
	g := rg.Group("/orders", middleware.Auth(verifier))
	if rl != nil {
		g.Use(rl)
	}
	g.POST("", h.Checkout)
	g.GET("/:id", h.Get)
	g.POST("/:id/cancel", h.Cancel)
}
