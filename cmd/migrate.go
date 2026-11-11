package cmd

import (
	"fmt"

	"github.com/alonespig/go-start/internal/infra/db"
	"github.com/spf13/cobra"
)

var (
	migrateDown   bool
	migrateStatus bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "执行数据库迁移",
	Long:  "执行版本化 SQL 迁移（internal/infra/db/migrations），版本记录在 goose_db_version 表。",
	RunE: func(c *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		// 迁移只需要 MySQL，直接用 db.New，不依赖 Redis
		gdb, err := db.New(cfg.MySQL)
		if err != nil {
			return err
		}
		sqlDB, err := gdb.DB()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		switch {
		case migrateStatus:
			return db.MigrateStatus(sqlDB)
		case migrateDown:
			return db.MigrateDown(sqlDB)
		default:
			if err := db.Migrate(sqlDB); err != nil {
				return fmt.Errorf("迁移失败: %w", err)
			}
			fmt.Println("迁移完成")
			return nil
		}
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&migrateDown, "down", false, "回滚最近一次迁移")
	migrateCmd.Flags().BoolVar(&migrateStatus, "status", false, "查看迁移状态")
}
