package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 应用指标集合，注册到默认 Prometheus registry（挂在 /metrics 上）
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge
}

func New() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数",
		}, []string{"method", "route", "status"}),
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求耗时（秒）",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		RequestsInFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "正在处理的 HTTP 请求数",
		}),
	}
}

// RegisterDBStats 暴露数据库连接池指标，便于观察连接泄漏与等待情况
func (m *Metrics) RegisterDBStats(db *sql.DB) {
	gauges := map[string]prometheus.GaugeFunc{
		"go_starter_db_open_connections": prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "go_starter_db_open_connections",
			Help: "当前打开的数据库连接数",
		}, func() float64 { return float64(db.Stats().OpenConnections) }),
		"go_starter_db_in_use_connections": prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "go_starter_db_in_use_connections",
			Help: "正在使用中的数据库连接数",
		}, func() float64 { return float64(db.Stats().InUse) }),
		"go_starter_db_idle_connections": prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "go_starter_db_idle_connections",
			Help: "空闲的数据库连接数",
		}, func() float64 { return float64(db.Stats().Idle) }),
		"go_starter_db_wait_count": prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "go_starter_db_wait_count",
			Help: "等待数据库连接的累计次数",
		}, func() float64 { return float64(db.Stats().WaitCount) }),
	}
	for _, g := range gauges {
		prometheus.MustRegister(g)
	}
}
