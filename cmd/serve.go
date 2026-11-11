package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alonespig/go-start/internal/di"
	"github.com/alonespig/go-start/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务",
	RunE: func(c *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
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
			// 只限制请求头读取，防慢连接；SSE / WebSocket 不受影响
			ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
			// 读请求体的总时限；长连接读（WebSocket）建议配 0
			ReadTimeout: cfg.Server.ReadTimeout,
			// 写响应的总时限；SSE 长连接建议配 0，或用 ResponseController 单独取消
			WriteTimeout: cfg.Server.WriteTimeout,
			// keep-alive 空闲回收；只作用于请求之间，不影响进行中的长连接
			IdleTimeout: cfg.Server.IdleTimeout,
			// 请求头大小上限，防止超大头部占用内存
			MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
		}

		// 监听失败（如端口被占用）要传回主 goroutine，否则进程会一直阻塞在 <-quit 上假装在跑
		errCh := make(chan error, 1)
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}()
		log.Info("http server started", zap.String("addr", addr))

		// 优雅退出
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-errCh:
			return fmt.Errorf("http server 启动失败: %w", err)
		case <-quit:
		}

		log.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}
