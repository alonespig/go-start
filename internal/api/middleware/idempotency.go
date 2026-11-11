package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/alonespig/go-start/internal/api/response"
	"github.com/alonespig/go-start/pkg/errcode"
)

// 幂等响应最多缓存 1MB，避免把大响应塞进 Redis
const maxIdempotentBody = 1 << 20

type cachedResponse struct {
	Status      int    `json:"status"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
}

// Idempotency 基于 Idempotency-Key 的请求去重：
// 同一 key 的并发请求只放行一个，完成后按原状态码和响应体重放。
// 客户端带同一个 key 重试时不会重复执行写操作。
func Idempotency(rdb *redis.Client, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || ttl <= 0 {
			c.Next()
			return
		}
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}
		if len(key) > 128 {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				response.Error(errcode.CodeInvalidParam, "Idempotency-Key 过长"))
			return
		}

		cacheKey := "idem:" + c.Request.Method + ":" + c.Request.URL.Path + ":" + key
		ctx := c.Request.Context()

		// 已完成过：按原响应重放
		if raw, err := rdb.Get(ctx, cacheKey).Bytes(); err == nil {
			var cached cachedResponse
			if json.Unmarshal(raw, &cached) == nil {
				c.Header("Idempotent-Replayed", "true")
				c.Data(cached.Status, cached.ContentType, []byte(cached.Body))
				c.Abort()
				return
			}
		}

		// 抢占处理权；抢不到说明同一 key 正在处理
		locked, err := rdb.SetNX(ctx, cacheKey, "processing", ttl).Result()
		if err == nil && !locked {
			c.AbortWithStatusJSON(http.StatusConflict,
				response.Error(errcode.CodeConflict, "相同 Idempotency-Key 的请求正在处理中"))
			return
		}

		writer := &bodyRecorder{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = writer
		c.Next()

		// 服务端错误允许客户端用同一个 key 重试
		if writer.status >= http.StatusInternalServerError || writer.body.Len() > maxIdempotentBody {
			_ = rdb.Del(context.WithoutCancel(ctx), cacheKey).Err()
			return
		}

		payload, err := json.Marshal(cachedResponse{
			Status:      writer.status,
			Body:        writer.body.String(),
			ContentType: writer.Header().Get("Content-Type"),
		})
		if err == nil {
			_ = rdb.Set(context.WithoutCancel(ctx), cacheKey, payload, ttl).Err()
		}
	}
}

// bodyRecorder 记录响应状态码与响应体，供幂等重放使用
type bodyRecorder struct {
	gin.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *bodyRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyRecorder) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyRecorder) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyRecorder) Status() int { return w.status }

func (w *bodyRecorder) Size() int { return w.body.Len() }
