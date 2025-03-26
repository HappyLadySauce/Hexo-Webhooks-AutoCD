package router

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/webhooks"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()
	// 注册 webhook 路由
	r.POST(config.Config.Webhook.Path, webhooks.HandleWebhook)
	return r
}
