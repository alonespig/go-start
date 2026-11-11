package cmd

import (
	"fmt"

	"github.com/alonespig/go-start/internal/infra/db"
	"github.com/alonespig/go-start/internal/model"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "写入开发用示例数据",
	RunE: func(c *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		gdb, err := db.New(cfg.MySQL)
		if err != nil {
			return err
		}
		defer func() {
			if sqlDB, err := gdb.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}()

		users := []model.User{
			{Name: "Demo", Email: "demo@example.com", Age: 20},
			{Name: "Alice", Email: "alice@example.com", Age: 30},
		}
		for i := range users {
			var count int64
			if err := gdb.Model(&model.User{}).Where("email = ?", users[i].Email).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			if err := gdb.Create(&users[i]).Error; err != nil {
				return err
			}
			fmt.Printf("seed: %s\n", users[i].Email)
		}
		return nil
	},
}
