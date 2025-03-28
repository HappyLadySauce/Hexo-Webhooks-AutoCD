package main

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/logger"
	"Hexo-AutoCD/router"
	"fmt"
	"os"
)

func main() {
	// 加载配置
	config.InitConfig()

	// 初始化日志系统
	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化路由
	r := router.InitRouter()

	// 根据配置决定使用 HTTP 还是 HTTPS
	addr := fmt.Sprintf(":%d", config.Config.Webhook.Port)

	// 日志记录启动信息
	logger.Info("服务器开始启动")

	if config.Config.SSL.Enabled {
		// 使用 HTTPS
		logger.Infof("HTTPS 服务器启动于 %s", addr)
		err := r.RunTLS(addr,
			config.Config.SSL.CertFile,
			config.Config.SSL.KeyFile)
		if err != nil {
			logger.Fatalf("启动 HTTPS 服务器失败: %v", err)
		}
	} else {
		// 使用 HTTP
		logger.Infof("HTTP 服务器启动于 %s", addr)
		err := r.Run(addr)
		if err != nil {
			logger.Fatalf("启动 HTTP 服务器失败: %v", err)
		}
	}
}
