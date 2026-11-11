package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/alonespig/go-start/internal/api/response"
	"github.com/alonespig/go-start/pkg/errcode"
)

// RateLimit 基于 Redis 的固定窗口限流：按 路由 + 客户端 IP 计数。
// Redis 不可用时放行（fail-open），避免缓存故障拖垮整个服务。
func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || limit <= 0 || window <= 0 {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		key := "ratelimit:" + c.FullPath() + ":" + c.ClientIP()

		n, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if n == 1 {
			_ = rdb.Expire(ctx, key, window).Err()
		}

		remaining := int64(limit) - n
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if n > int64(limit) {
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				response.Error(errcode.CodeTooManyRequests, "请求过于频繁，请稍后再试"))
			return
		}
		c.Next()
	}
}
