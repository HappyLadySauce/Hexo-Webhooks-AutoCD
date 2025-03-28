package middlewares

import (
	"Hexo-AutoCD/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DenyScan() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 如果请求路径不是/webhooks，则返回错误信息和IP地址
		if c.Request.URL.Path != "/webhook" {
			logger.Warnf("检测到扫描请求: %s %s 来自 %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
			c.JSON(http.StatusNotFound, gin.H{"message": "请不要扫描我的博客！", "ip": c.ClientIP()})
			return
		}
		// 记录请求日志
		logger.Infof("接收到Webhook请求: %s %s 来自 %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	})
}
