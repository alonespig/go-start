package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/alonespig/go-start/internal/api/response"
	"github.com/alonespig/go-start/internal/auth"
	"github.com/alonespig/go-start/pkg/errcode"
)

// AuthClaimsKey 鉴权成功后 claims 在 gin.Context 中的键
const AuthClaimsKey = "auth_claims"

type claimsCtxKey struct{}

// Auth 校验 Authorization: Bearer <token>，成功后把 claims 放进 context
func Auth(ts *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			abortUnauthorized(c)
			return
		}

		claims, err := ts.Parse(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(AuthClaimsKey, claims)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), claimsCtxKey{}, claims),
		)
		c.Next()
	}
}

// ClaimsFromGin 从 gin.Context 取 claims（handler 使用）
func ClaimsFromGin(c *gin.Context) *auth.Claims {
	v, ok := c.Get(AuthClaimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*auth.Claims)
	return claims
}

// ClaimsFromContext 从 request context 取 claims（service 使用）
func ClaimsFromContext(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsCtxKey{}).(*auth.Claims)
	return claims
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized,
		response.Error(errcode.CodeUnauthorized, "未登录或 token 已失效"))
}
