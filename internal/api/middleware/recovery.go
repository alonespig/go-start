package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-starter/pkg/errcode"
	"go-starter/pkg/log"
	"go.uber.org/zap"
)

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", zap.Any("err", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    errcode.CodeInternal,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
