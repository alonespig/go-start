package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version 可通过 -ldflags "-X go-starter/cmd.version=x.y.z" 注入
var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本号",
	Run: func(c *cobra.Command, args []string) {
		fmt.Printf("go-starter version %s\n", version)
	},
}
