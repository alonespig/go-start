package di

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go-starter/config"
	"go-starter/internal/api/handler"
	"go-starter/internal/api/router"
	"go-starter/internal/infra/db"
	infraMongo "go-starter/internal/infra/mongo"
	infraRedis "go-starter/internal/infra/redis"
	"go-starter/internal/repository"
	"go-starter/internal/service"
	"go-starter/pkg/log"
)

// Container 组合根：持有装配好的所有依赖
type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Mongo  *mongo.Client

	UserService service.UserService
	Router      *gin.Engine
}

// Build 按依赖顺序装配整个对象图：log → infra → repository → service → api
func Build(cfg *config.Config) (*Container, error) {
	// 1. 用配置覆盖全局 logger 的默认初始化（init 里已用默认值初始化过）
	if err := log.Init(cfg.Log); err != nil {
		return nil, err
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

	// Mongo 可选：连不上只告警（当前尚未使用）
	var mdb *mongo.Client
	if mdb, err = infraMongo.New(cfg.Mongo); err != nil {
		log.Warn("mongo 连接失败，降级为不可用", zap.Error(err))
		mdb = nil
	}

	// 3. repository（注入连接，返回接口）
	userRepo := repository.NewUserRepository(gdb)

	// 4. service（注入 repository，返回接口）
	userSvc := service.NewUserService(userRepo, rdb)

	// 5. handler（注入 service）
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userSvc)

	// 6. router（注入 handler；logger 走全局，不再传）
	r := router.New(healthHandler, userHandler)

	return &Container{
		Config:      cfg,
		DB:          gdb,
		Redis:       rdb,
		Mongo:       mdb,
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
	if c.Mongo != nil {
		_ = c.Mongo.Disconnect(context.Background())
	}
}
