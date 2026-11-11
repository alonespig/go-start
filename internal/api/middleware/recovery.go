package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/alonespig/go-start/pkg/errcode"
	"github.com/alonespig/go-start/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("err", r),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    errcode.CodeInternal,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
