package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/logger"
	"github.com/quangdangfit/goticket/pkg/response"
)

// TokenVerifier validates a bearer token and returns (userID, role).
type TokenVerifier interface {
	Verify(token string) (userID, role string, err error)
}

// Auth requires a valid bearer token. The decoded user_id is stored on ctx.
func Auth(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		uid, role, err := v.Verify(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set("user_id", uid)
		c.Set("role", role)
		c.Request = c.Request.WithContext(logger.WithUserID(c.Request.Context(), uid))
		c.Next()
	}
}

// RequireRole gates a route by role string.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r, _ := c.Get("role"); r != role {
			response.Fail(c, apperr.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
