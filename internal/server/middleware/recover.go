package middleware

import (
	"github.com/gin-gonic/gin"

	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/logger"
	"github.com/quangdangfit/goticket/pkg/response"
)

// Recover converts panics into a 500 envelope.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(c.Request.Context()).Error("panic recovered", "panic", r)
				response.Fail(c, apperr.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}
