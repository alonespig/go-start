package main

import (
	"go-starter/cmd"

	// 导入即用默认配置初始化全局 logger
	_ "go-starter/pkg/log"
	"go-starter/pkg/validation"
)

func main() {
	// 把校验错误里的字段名映射成 JSON 名称，启动时注册一次
	validation.Init()
	cmd.Execute()
}
