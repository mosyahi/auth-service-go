package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "limit:" + c.ClientIP()
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "redis error"})
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, time.Minute)
		}

		if count > 5 { // Maksimal 5 kali percobaan per menit
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
