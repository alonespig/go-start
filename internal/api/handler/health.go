package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/alonespig/go-start/internal/api/response"
)

// HealthHandler 健康检查：/health 为存活探针，/ready 为就绪探针
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: rdb}
}

// Check 存活探针：进程能响应即视为存活
func (h *HealthHandler) Check(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

// Ready 就绪探针：检查 MySQL / Redis 是否可用。
// Redis 是可选依赖，未启用时标记 disabled，不影响就绪判断。
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string, 2)
	ready := true

	if sqlDB, err := h.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		checks["mysql"] = "down"
		ready = false
	} else {
		checks["mysql"] = "ok"
	}

	if h.redis == nil {
		checks["redis"] = "disabled"
	} else if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "down"
		ready = false
	} else {
		checks["redis"] = "ok"
	}

	status := http.StatusOK
	body := gin.H{"status": "ok", "checks": checks}
	if !ready {
		status = http.StatusServiceUnavailable
		body["status"] = "degraded"
	}
	c.JSON(status, body)
}
