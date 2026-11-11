package di

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/alonespig/go-start/config"
	"github.com/alonespig/go-start/internal/api/handler"
	"github.com/alonespig/go-start/internal/api/router"
	"github.com/alonespig/go-start/internal/infra/db"
	infraRedis "github.com/alonespig/go-start/internal/infra/redis"
	"github.com/alonespig/go-start/internal/repository"
	"github.com/alonespig/go-start/internal/service"
	"github.com/alonespig/go-start/pkg/log"
)

// Container 组合根：持有装配好的所有依赖
type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client

	UserService service.UserService
	Router      *gin.Engine
}

// Build 按依赖顺序装配整个对象图：log → infra → repository → service → api
func Build(cfg *config.Config) (*Container, error) {
	// 1. 用配置覆盖全局 logger 的默认初始化（init 里已用默认值初始化过）
	if err := log.Init(cfg.Log); err != nil {
		return nil, err
	}

	// 生产环境关闭 gin debug 模式（路由打印、额外检查），必须在 router.New 之前设置
	if cfg.App.Env == "prod" || cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. infra：MySQL 是强依赖
	gdb, err := db.New(cfg.MySQL)
	if err != nil {
		return nil, err
	}

	// Redis 可选：连不上只告警，降级为无缓存（GetByID 自动退化查库）
	var rdb *redis.Client
	if rdb, err = infraRedis.New(cfg.Redis); err != nil {
		log.Warn("redis 连接失败，降级为无缓存", zap.Error(err))
		rdb = nil
	}

	// 3. repository（注入连接，返回接口）
	userRepo := repository.NewUserRepository(gdb)

	// 4. service（注入 repository，返回接口）
	userSvc := service.NewUserService(userRepo, rdb)

	// 5. handler（注入 service / 依赖，供探针使用）
	healthHandler := handler.NewHealthHandler(gdb, rdb)
	userHandler := handler.NewUserHandler(userSvc)

	// 6. router（注入配置与 handler；logger 走全局，不再传）
	r, err := router.New(cfg, healthHandler, userHandler)
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:      cfg,
		DB:          gdb,
		Redis:       rdb,
		UserService: userSvc,
		Router:      r,
	}, nil
}

// Close 释放资源
func (c *Container) Close() {
	if c.DB != nil {
		if sqlDB, err := c.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
}
