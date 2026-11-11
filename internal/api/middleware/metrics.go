package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alonespig/go-start/internal/infra/metrics"
)

// Metrics 采集 HTTP 请求量、状态码与耗时；route 用路由模板避免标签基数爆炸
func Metrics(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.RequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		m.RequestsInFlight.Dec()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		m.RequestsTotal.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.RequestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}
