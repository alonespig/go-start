package cmd

import (
	"fmt"
	"os"

	"github.com/alonespig/go-start/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "go-starter",
	Short: "go-starter 通用后端脚手架",
	Long:  "一个基于 Gin + GORM + Redis 的通用后端脚手架，支持多命令启动。",
}

// Execute 是 CLI 入口，由 main 调用
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// loadConfig 统一加载配置并补上命令层错误前缀，避免各命令重复处理
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	return cfg, nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "配置文件路径")
	rootCmd.AddCommand(serveCmd, cronCmd, migrateCmd, seedCmd, versionCmd)
}
