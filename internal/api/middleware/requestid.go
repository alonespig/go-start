package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "request_id"

type requestIDCtxKey struct{}

// RequestID 读取或生成 X-Request-ID，并写入响应头、gin.Context 与 request context，
// 便于把一次请求的日志串起来（后续接入链路追踪时也可复用）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), requestIDCtxKey{}, id),
		)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestIDFromGin 返回当前请求的 ID（中间件里使用）
func RequestIDFromGin(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

// RequestIDFromContext 供 service / repository 打日志时取请求 ID
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDCtxKey{}).(string)
	return id
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
