package handler

import (
	"github.com/gin-gonic/gin"
	"go-starter/internal/api/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 健康检查
func (h *HealthHandler) Check(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
