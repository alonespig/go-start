package db

import (
	"fmt"
	"time"

	"go-starter/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// New 初始化 GORM MySQL 连接
func New(cfg config.MySQLConfig) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		// SQL 日志接入全局 zap（普通 SQL 为 debug 级别，慢查询/错误会打到 warn/error）
		Logger: NewGormLogger(),
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
	sqlDB.SetConnMaxLifetime(time.Hour)

	return gdb, nil
}
