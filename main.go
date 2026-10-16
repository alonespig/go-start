package main

import (
	"go-starter/cmd"

	// 导入即用默认配置初始化全局 logger
	_ "go-starter/pkg/log"
)

func main() {
	cmd.Execute()
}
