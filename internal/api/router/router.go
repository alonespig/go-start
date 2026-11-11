package router

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/alonespig/go-start/config"
	"github.com/alonespig/go-start/internal/api/handler"
	"github.com/alonespig/go-start/internal/api/middleware"
)

// New 组装路由，注入配置与 handler
func New(cfg *config.Config, health *handler.HealthHandler, user *handler.UserHandler) (*gin.Engine, error) {
	r := gin.New()

	// 不信任任何代理头（nil / 空列表）时，ClientIP 取直连地址，避免伪造 X-Forwarded-For
	if err := r.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("server.trusted_proxies 配置非法: %w", err)
	}

	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(cfg.CORS),
	)

	// 探针：/health 存活，/ready 检查依赖
	r.GET("/health", health.Check)
	r.GET("/ready", health.Ready)

	// 业务路由（版本化，便于后续并行维护 v2）
	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", user.Create)
			users.GET("/:id", user.Get)
			users.PUT("/:id", user.Update)
			users.DELETE("/:id", user.Delete)
			users.GET("", user.List)
		}
	}
	return r, nil
}
