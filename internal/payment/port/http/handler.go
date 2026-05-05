// Package http exposes the payment webhook over Gin.
package http

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/internal/payment"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
)

// Handler exposes the webhook entrypoint.
type Handler struct{ svc payment.Service }

// NewHandler constructs a Handler.
func NewHandler(svc payment.Service) *Handler { return &Handler{svc: svc} }

// Webhook handles POST /webhook/payment/:provider — body is forwarded to
// the gateway-specific verifier in the service.
func (h *Handler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, apperr.ErrInvalidPayload)
		return
	}
	headers := map[string]string{}
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	if err := h.svc.HandleWebhook(c.Request.Context(), headers, body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}
