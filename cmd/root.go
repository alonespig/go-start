package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "go-starter",
	Short: "go-starter 通用后端脚手架",
	Long:  "一个基于 Gin + GORM + Redis + MongoDB 的通用后端脚手架，支持多命令启动。",
}

// Execute 是 CLI 入口，由 main 调用
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "配置文件路径")
	rootCmd.AddCommand(serveCmd, cronCmd, migrateCmd, versionCmd)
}
