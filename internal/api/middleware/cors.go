package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/alonespig/go-start/config"
)

// CORS 跨域中间件：允许来源与凭证由配置驱动。
// 携带凭证（Cookie / Authorization）时不能返回 "*"，必须回显具体 Origin。
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowAll := len(cfg.AllowedOrigins) == 0 || slices.Contains(cfg.AllowedOrigins, "*")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		switch {
		case allowAll && cfg.AllowCredentials && origin != "":
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		case allowAll:
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "" && slices.Contains(cfg.AllowedOrigins, origin):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
