package webhooks

import (
	"Hexo-AutoCD/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"


	"github.com/gin-gonic/gin"
)

// Github 的 signature = "sha256=" + HMAC-SHA256(secret, body),使用 HMAC-SHA256 算法验证请求的签名
func verifySignature(signature string, body []byte, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return signature == expectedSignature
}

func HandleWebhook(c *gin.Context) {
	// 获取请求头中的 signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"错误": "X-Hub-Signature-256 头缺失"})
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"错误": "读取请求体失败"})
		return
	}

	// 使用 body 和 secret 验证请求签名(signature)
	isValid := verifySignature(signature, body, config.Config.Webhook.Secret)
	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"错误": "X-Hub-Signature-256 头不匹配"})
		return
	}

	// 获取请求体中的事件类型
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"错误": "X-GitHub-Event 头缺失"})
		return
	}

	// 根据事件类型进行不同的处理
	switch eventType {
	case "push":
		handlePushEvent(c, body)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"错误": "不支持的事件类型"})
	}
}

func handlePushEvent(c *gin.Context, body []byte) {
	
}