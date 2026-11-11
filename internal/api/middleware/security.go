package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders 常见安全响应头；HSTS 只在 HTTPS 下有意义，按需自行添加
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
