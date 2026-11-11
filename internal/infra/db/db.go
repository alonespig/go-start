package db

import (
	"fmt"

	"github.com/alonespig/go-start/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// New 初始化 GORM MySQL 连接
func New(cfg config.MySQLConfig) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		// SQL 日志接入全局 zap（普通 SQL 为 debug 级别，慢查询/错误会打到 warn/error）
		Logger: NewGormLogger(cfg.SlowThreshold),
		// 把驱动错误翻译成 gorm.ErrDuplicatedKey 等通用错误，便于上层统一处理
		TranslateError: true,
		// 复用预编译语句，减少 SQL 解析开销；部分代理（如某些中间件）可能不支持
		PrepareStmt: cfg.PrepareStmt,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 mysql 失败: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return gdb, nil
}
