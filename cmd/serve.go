package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go-starter/config"
	"go-starter/internal/di"
	"go-starter/pkg/log"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务",
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

		addr := fmt.Sprintf(":%d", cfg.App.Port)
		srv := &http.Server{
			Addr:    addr,
			Handler: container.Router,
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("http server error", zap.Error(err))
			}
		}()
		log.Info("http server started", zap.String("addr", addr))

		// 优雅退出
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}
