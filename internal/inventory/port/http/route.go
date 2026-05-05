package http

import (
	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/server/middleware"
)

// RegisterRoutes mounts /holds endpoints (auth required).
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, verifier middleware.TokenVerifier) {
	g := rg.Group("/holds", middleware.Auth(verifier))
	g.POST("", h.CreateHold)
	g.DELETE("/:id", h.DeleteHold)
}
