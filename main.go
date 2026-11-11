package main

import (
	"github.com/alonespig/go-start/cmd"
	"github.com/alonespig/go-start/internal/api/validation"

	// 导入即用默认配置初始化全局 logger
	_ "github.com/alonespig/go-start/pkg/log"
)

func main() {
	// 把校验错误里的字段名映射成 JSON 名称，启动时注册一次
	validation.Init()
	cmd.Execute()
}
