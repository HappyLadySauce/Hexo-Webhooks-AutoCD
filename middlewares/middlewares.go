package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

func DenyScan() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 如果请求路径不是/webhooks，则返回错误信息和IP地址
		if c.Request.URL.Path != "/webhook" {
			c.JSON(http.StatusNotFound, gin.H{"message": "请不要扫描我的博客！", "ip": c.ClientIP()})
			return
		}
		// 记录请求日志
		log.Printf("Request: %s %s %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	})
}
