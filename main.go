package main

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/router"
	"fmt"
	"log"
)

func main() {
	// 加载配置
	config.InitConfig()

	// 初始化路由
	r := router.InitRouter()

	// 根据配置决定使用 HTTP 还是 HTTPS
	addr := fmt.Sprintf(":%d", config.Config.Webhook.Port)
	if config.Config.SSL.Enabled {
		// 使用 HTTPS
		err := r.RunTLS(addr, 
			config.Config.SSL.CertFile,
			config.Config.SSL.KeyFile)
		if err != nil {
			log.Fatalf("启动 HTTPS 服务器失败: %v", err)
		}
	} else {
		// 使用 HTTP
		err := r.Run(addr)
		if err != nil {
			log.Fatalf("启动 HTTP 服务器失败: %v", err)
		}
	}
}
