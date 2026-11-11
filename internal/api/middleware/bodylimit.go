package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit 限制请求体大小，超出后读取会返回错误（由参数校验统一处理）
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
