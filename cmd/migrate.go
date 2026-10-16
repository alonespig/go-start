package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go-starter/config"
	"go-starter/internal/infra/db"
	"go-starter/internal/model"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "执行数据库迁移",
	RunE: func(c *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 迁移只需要 MySQL，直接用 db.New，不依赖 Redis/Mongo
		gdb, err := db.New(cfg.MySQL)
		if err != nil {
			return err
		}

		// 迁移表结构：在这里追加所有需要迁移的 model
		if err := gdb.AutoMigrate(&model.User{}); err != nil {
			return fmt.Errorf("迁移失败: %w", err)
		}
		fmt.Println("迁移完成")
		return nil
	},
}
