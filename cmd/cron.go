package cmd

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go-starter/config"
	"go-starter/internal/di"
	"go-starter/pkg/log"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "启动定时任务",
	RunE: func(c *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		container, err := di.Build(cfg)
		if err != nil {
			return err
		}
		defer container.Close()
		defer func() { _ = log.L().Sync() }()

		cj := cron.New()
		// TODO: 在这里注册定时任务，例如：
		// cj.AddFunc("@every 1m", func() { log.Info("tick") })

		log.Info("cron scheduler started")
		cj.Run() // 阻塞运行
		return nil
	},
}
