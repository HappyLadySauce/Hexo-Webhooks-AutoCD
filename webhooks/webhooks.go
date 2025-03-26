package webhooks

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/scripts"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Github 的 signature = "sha256=" + HMAC-SHA256(secret, body)
func verifySignature(signature string, body []byte, secret string) bool {
	// 检查前缀
	const prefix = "sha256="
	if len(signature) <= len(prefix) || signature[:len(prefix)] != prefix {
		return false
	}

	// 去除 "sha256=" 前缀
	signature = signature[len(prefix):]

	// 计算 HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return signature == expectedSignature
}

func HandleWebhook(c *gin.Context) {
	// 获取请求头中的 signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		log.Printf("错误：缺少 X-Hub-Signature-256 头")
		c.JSON(http.StatusBadRequest, gin.H{"错误": "X-Hub-Signature-256 头缺失"})
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("错误：读取请求体失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"错误": "读取请求体失败"})
		return
	}

	// 验证签名
	isValid := verifySignature(signature, body, config.Config.Webhook.Secret)
	if !isValid {
		log.Printf("错误：签名验证失败。收到的签名：%s", signature)
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
		handlePushEvent(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"错误": "不支持的事件类型"})
	}
}

func handlePushEvent(c *gin.Context) {
	// 执行脚本
	executor := scripts.NewExecutor(scripts.ExecutorConfig{
		ScriptsPath:   config.Config.Scripts.Path,
		Timeout:       30 * time.Second,
		MaxConcurrent: 5,
	})
	result, err := executor.Execute(config.Config.Scripts.Push, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"错误": err.Error()})
		return
	}
	if result.ExitCode != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"错误": result.Output})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{"消息": "脚本执行成功", "结果": result.Output})
}
