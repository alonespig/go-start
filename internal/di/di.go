package di

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/alonespig/go-start/config"
	"github.com/alonespig/go-start/internal/api/handler"
	"github.com/alonespig/go-start/internal/api/middleware"
	"github.com/alonespig/go-start/internal/api/router"
	"github.com/alonespig/go-start/internal/auth"
	"github.com/alonespig/go-start/internal/infra/db"
	kafka "github.com/alonespig/go-start/internal/infra/kafka"
	"github.com/alonespig/go-start/internal/infra/metrics"
	infraRedis "github.com/alonespig/go-start/internal/infra/redis"
	"github.com/alonespig/go-start/internal/repository"
	"github.com/alonespig/go-start/internal/service"
	"github.com/alonespig/go-start/pkg/cache"
	"github.com/alonespig/go-start/pkg/log"
)

// Container 组合根：持有装配好的所有依赖
type Container struct {
	Config        *config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	Cache         cache.Cache
	TxManager     *repository.TxManager
	TokenService  *auth.TokenService
	Metrics       *metrics.Metrics
	KafkaProducer kafka.Producer

	UserService service.UserService
	Router      *gin.Engine
}

// Build 按依赖顺序装配整个对象图：log → infra → repository → service → api
func Build(cfg *config.Config) (*Container, error) {
	// 1. 用配置覆盖全局 logger 的默认初始化（init 里已用默认值初始化过）
	if err := log.Init(log.Options{
		Level:      cfg.Log.Level,
		File:       cfg.Log.File,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}); err != nil {
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

	// 指标（可选）：注册 HTTP 与数据库连接池指标
	var appMetrics *metrics.Metrics
	if cfg.Metrics.Enabled {
		appMetrics = metrics.New()
		if sqlDB, err := gdb.DB(); err == nil {
			appMetrics.RegisterDBStats(sqlDB)
		}
	}

	// Redis 可选：连不上只告警，降级为无缓存（GetByID 自动退化查库）
	var rdb *redis.Client
	if rdb, err = infraRedis.New(cfg.Redis); err != nil {
		log.Warn("redis 连接失败，降级为无缓存", zap.Error(err))
		rdb = nil
	}

	var userCache cache.Cache
	if rdb != nil {
		userCache = infraRedis.NewCache(rdb)
	}

	// Kafka 可选：默认关闭，开启后提供 Producer，Consumer 按 topic 创建
	var kafkaProducer kafka.Producer
	if cfg.Kafka.Enabled {
		kafkaProducer, err = kafka.NewProducer(cfg.Kafka)
		if err != nil {
			return nil, fmt.Errorf("初始化 kafka 失败: %w", err)
		}
	}

	// 3. repository（注入连接，返回接口）
	userRepo := repository.NewUserRepository(gdb)
	txManager := repository.NewTxManager(gdb)

	// 4. service（注入 repository，返回接口）
	userSvc := service.NewUserService(userRepo, userCache)

	// 5. 鉴权（可选）：配置了 jwt_secret 才启用，并注册受保护路由
	var tokenSvc *auth.TokenService
	var authMW gin.HandlerFunc
	if cfg.Auth.Enabled() {
		tokenSvc, err = auth.NewTokenService(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.JWTExpire)
		if err != nil {
			return nil, fmt.Errorf("初始化鉴权失败: %w", err)
		}
		authMW = middleware.Auth(tokenSvc)
	}

	// 6. handler（注入 service / 依赖，供探针使用）
	healthHandler := handler.NewHealthHandler(gdb, rdb)
	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler()

	// 7. router（注入配置与 handler；logger 走全局，不再传）
	r, err := router.New(router.Deps{
		Config:         cfg,
		Health:         healthHandler,
		User:           userHandler,
		Auth:           authHandler,
		AuthMiddleware: authMW,
		Metrics:        appMetrics,
		Redis:          rdb,
	})
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:        cfg,
		DB:            gdb,
		Redis:         rdb,
		Cache:         userCache,
		TxManager:     txManager,
		TokenService:  tokenSvc,
		Metrics:       appMetrics,
		KafkaProducer: kafkaProducer,
		UserService:   userSvc,
		Router:        r,
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
	if c.KafkaProducer != nil {
		_ = c.KafkaProducer.Close()
	}
}
