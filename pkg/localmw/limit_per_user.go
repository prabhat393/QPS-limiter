package localmw

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// LimitPerUser middleware is used to limit number of request per second for user.
// Note: if expiration is set to 0 it means the key has no expiration time.
func LimitPerUser(cmdable redis.Cmdable, limit int, key string, expiration time.Duration) gin.HandlerFunc {
	limiter := NewRedisLimiter(cmdable, key, limit, expiration)

	return func(c *gin.Context) {

		allowed, err := limiter.Allow(c.Request.Context(), key)

		if err != nil {
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		if !allowed {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		if err := limiter.Seen(c.Request.Context(), key); err != nil {
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Next()
	}
}
