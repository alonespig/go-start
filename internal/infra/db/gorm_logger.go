package db

import (
	"context"
	"errors"
	"time"

	"github.com/alonespig/go-start/pkg/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormLogger 把 GORM 的 SQL 日志接到全局 zap，统一日志格式。
// 普通 SQL 走 debug 级别（想看 SQL 把 log.level 配成 debug），慢查询 warn，错误 error。
type gormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger 返回接入全局 zap 的 GORM logger
func NewGormLogger(slowThreshold time.Duration) gormlogger.Interface {
	if slowThreshold <= 0 {
		slowThreshold = 200 * time.Millisecond
	}
	return &gormLogger{level: gormlogger.Info, slowThreshold: slowThreshold}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		log.Info(msg, zap.Any("data", data))
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		log.Warn(msg, zap.Any("data", data))
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		log.Error(msg, zap.Any("data", data))
	}
}

// Trace 记录 SQL、耗时、影响行数；区分错误 / 慢查询 / 普通
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		log.Error("gorm error",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("latency", elapsed),
		)
	case elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		log.Warn("gorm slow sql",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("latency", elapsed),
		)
	case l.level >= gormlogger.Info:
		log.Debug("gorm sql",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("latency", elapsed),
		)
	}
}
