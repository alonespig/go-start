package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/alonespig/go-start/internal/api/middleware"
	"github.com/alonespig/go-start/internal/api/response"
	"github.com/alonespig/go-start/pkg/errcode"
)

// AuthHandler 鉴权相关接口；登录逻辑接入后复用 auth.TokenService.Issue 签发 token。
type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Me 返回当前 token 的身份信息，用于验证鉴权链路
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.ClaimsFromGin(c)
	if claims == nil {
		response.Fail(c, errcode.New(errcode.CodeUnauthorized, "未登录"))
		return
	}
	response.OK(c, gin.H{"user_id": claims.UserID, "roles": claims.Roles})
}
