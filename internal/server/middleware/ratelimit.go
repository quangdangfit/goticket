package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/response"
)

// RateLimit applies a fixed-window per-key limit using Redis INCR + EXPIRE.
// keyFn picks the bucket key (e.g. user id, IP).
func RateLimit(rdb *redis.Client, perMinute int, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if perMinute <= 0 {
			c.Next()
			return
		}
		key := "rl:" + keyFn(c) + ":" + strconv.FormatInt(time.Now().Unix()/60, 10)
		ctx := c.Request.Context()
		n, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if n == 1 {
			_ = rdb.Expire(ctx, key, time.Minute).Err()
		}
		if n > int64(perMinute) {
			response.Fail(c, apperr.ErrRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
