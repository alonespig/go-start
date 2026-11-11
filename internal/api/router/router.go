package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/alonespig/go-start/config"
	"github.com/alonespig/go-start/internal/api/handler"
	"github.com/alonespig/go-start/internal/api/middleware"
	"github.com/alonespig/go-start/internal/infra/metrics"
)

// Deps 路由依赖；Auth 为 nil 时不注册受保护路由
type Deps struct {
	Config         *config.Config
	Health         *handler.HealthHandler
	User           *handler.UserHandler
	Auth           *handler.AuthHandler
	AuthMiddleware gin.HandlerFunc
	Metrics        *metrics.Metrics
	Redis          *redis.Client
}

// New 组装路由
func New(d Deps) (*gin.Engine, error) {
	r := gin.New()

	// 不信任任何代理头（nil / 空列表）时，ClientIP 取直连地址，避免伪造 X-Forwarded-For
	if err := r.SetTrustedProxies(d.Config.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("server.trusted_proxies 配置非法: %w", err)
	}

	handlers := []gin.HandlerFunc{
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(d.Config.CORS),
		middleware.SecurityHeaders(),
		middleware.BodyLimit(d.Config.Server.MaxBodyBytes),
	}
	if d.Metrics != nil {
		handlers = append(handlers, middleware.Metrics(d.Metrics))
	}
	if d.Config.RateLimit.Enabled {
		handlers = append(handlers, middleware.RateLimit(
			d.Redis, d.Config.RateLimit.Requests, d.Config.RateLimit.Window))
	}
	if d.Config.Idempotency.Enabled {
		handlers = append(handlers, middleware.Idempotency(d.Redis, d.Config.Idempotency.TTL))
	}
	r.Use(handlers...)

	// 探针：/health 存活，/ready 检查依赖
	r.GET("/health", d.Health.Check)
	r.GET("/ready", d.Health.Ready)

	// Prometheus 指标
	if d.Metrics != nil && d.Config.Metrics.Enabled {
		path := d.Config.Metrics.Path
		if path == "" {
			path = "/metrics"
		}
		r.GET(path, gin.WrapH(promhttp.Handler()))
	}

	// 业务路由（版本化，便于后续并行维护 v2）
	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", d.User.Create)
			users.GET("/:id", d.User.Get)
			users.PUT("/:id", d.User.Update)
			users.DELETE("/:id", d.User.Delete)
			users.GET("", d.User.List)
		}
	}

	// 受保护路由：启用鉴权时才注册
	if d.AuthMiddleware != nil && d.Auth != nil {
		authed := api.Group("", d.AuthMiddleware)
		{
			authed.GET("/me", d.Auth.Me)
		}
	}
	return r, nil
}
