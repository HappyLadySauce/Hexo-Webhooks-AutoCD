package main

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/router"
	"fmt"
)

func main() {
	// 加载配置
	config.InitConfig()

	// 初始化路由
	r := router.InitRouter()

	// 启动服务器
	r.Run(fmt.Sprintf(":%d", config.Config.Webhook.Port))
}
