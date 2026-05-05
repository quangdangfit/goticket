package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the provider webhook under /webhook.
// IMPORTANT: the route group must NOT have body-mutating middleware so the
// raw bytes used in HMAC verification stay intact.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	rg.POST("/payment/:provider", h.Webhook)
}
