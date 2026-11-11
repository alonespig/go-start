package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alonespig/go-start/config"
	"github.com/alonespig/go-start/internal/di"
	"github.com/alonespig/go-start/pkg/log"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
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

		cj.Start()
		log.Info("cron scheduler started")

		// 优雅退出：收到信号后等待运行中的任务结束，最多等 10s
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info("stopping cron scheduler...")
		done := cj.Stop()
		select {
		case <-done.Done():
		case <-time.After(10 * time.Second):
			log.Warn("cron 任务未在 10s 内结束")
		}
		return nil
	},
}
