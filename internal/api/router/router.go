package router

import (
	"github.com/gin-gonic/gin"
	"go-starter/internal/api/handler"
	"go-starter/internal/api/middleware"
)

// New 组装路由，注入 handler
func New(health *handler.HealthHandler, user *handler.UserHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())

	// 健康检查
	r.GET("/health", health.Check)

	// 业务路由
	api := r.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("", user.Create)
			users.GET("/:id", user.Get)
			users.GET("", user.List)
		}
	}
	return r
}
