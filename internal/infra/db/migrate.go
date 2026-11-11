package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate 执行所有未应用的迁移（版本记录在 goose_db_version 表）
func Migrate(sqlDB *sql.DB) error {
	if err := prepareGoose(); err != nil {
		return err
	}
	return goose.Up(sqlDB, "migrations")
}

// MigrateDown 回滚最近一次迁移
func MigrateDown(sqlDB *sql.DB) error {
	if err := prepareGoose(); err != nil {
		return err
	}
	return goose.Down(sqlDB, "migrations")
}

// MigrateStatus 打印迁移状态
func MigrateStatus(sqlDB *sql.DB) error {
	if err := prepareGoose(); err != nil {
		return err
	}
	return goose.Status(sqlDB, "migrations")
}

func prepareGoose() error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("初始化迁移工具失败: %w", err)
	}
	return nil
}
