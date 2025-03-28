package webhooks

import (
	"Hexo-AutoCD/config"
	"Hexo-AutoCD/logger"
	"Hexo-AutoCD/scripts"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		logger.Error("缺少 X-Hub-Signature-256 头")
		c.JSON(http.StatusBadRequest, gin.H{"错误": "X-Hub-Signature-256 头缺失"})
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Errorf("读取请求体失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"错误": "读取请求体失败"})
		return
	}

	// 验证签名
	isValid := verifySignature(signature, body, config.Config.Webhook.Secret)
	if !isValid {
		logger.Errorf("签名验证失败。收到的签名：%s", signature)
		c.JSON(http.StatusUnauthorized, gin.H{"错误": "X-Hub-Signature-256 头不匹配"})
		return
	}

	// 获取请求体中的事件类型
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType == "" {
		logger.Error("缺少 X-GitHub-Event 头")
		c.JSON(http.StatusBadRequest, gin.H{"错误": "X-GitHub-Event 头缺失"})
		return
	}

	logger.Infof("收到 GitHub %s 事件", eventType)

	// 根据事件类型进行不同的处理
	switch eventType {
	case "push":
		logger.WithField("事件类型", "push").Info("处理推送事件")
		handlePushEvent(c, body)
	default:
		logger.WithFields(logrus.Fields{
			"事件类型": eventType,
			"IP地址": c.ClientIP(),
		}).Warn("收到不支持的事件类型")
		c.JSON(http.StatusBadRequest, gin.H{"错误": "不支持的事件类型"})
	}
}

// Git commit
type HeadCommit struct {
	ID        string   `json:"id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

type PushEvent struct {
	HeadCommit HeadCommit `json:"head_commit"`
}

func handlePushEvent(c *gin.Context, body []byte) {
	// 解析 body
	var pushEvent PushEvent
	if err := json.Unmarshal(body, &pushEvent); err != nil {
		logger.WithError(err).Error("无法解析 push 事件数据")
		c.JSON(http.StatusBadRequest, gin.H{"错误": "无法解析 push 事件数据"})
		return
	}

	// 截取提交ID的前8位以便于显示
	shortCommitID := pushEvent.HeadCommit.ID
	if len(shortCommitID) > 8 {
		shortCommitID = shortCommitID[:8]
	}

	logger.WithFields(logrus.Fields{
		"提交ID":  shortCommitID,
		"提交信息":  pushEvent.HeadCommit.Message,
		"提交时间":  pushEvent.HeadCommit.Timestamp,
		"新增文件数": len(pushEvent.HeadCommit.Added),
		"修改文件数": len(pushEvent.HeadCommit.Modified),
		"删除文件数": len(pushEvent.HeadCommit.Removed),
	}).Info("收到Git推送事件")

	// 准备环境变量
	commitEnv := []string{
		fmt.Sprintf("COMMIT_ID=%s", pushEvent.HeadCommit.ID),
		fmt.Sprintf("COMMIT_MESSAGE=%s", pushEvent.HeadCommit.Message),
		fmt.Sprintf("COMMIT_TIMESTAMP=%s", pushEvent.HeadCommit.Timestamp),
		fmt.Sprintf("COMMIT_ADDED=%s", strings.Join(pushEvent.HeadCommit.Added, ",")),
		fmt.Sprintf("COMMIT_REMOVED=%s", strings.Join(pushEvent.HeadCommit.Removed, ",")),
		fmt.Sprintf("COMMIT_MODIFIED=%s", strings.Join(pushEvent.HeadCommit.Modified, ",")),
	}

	// 创建执行器
	executor := scripts.NewExecutor(scripts.ExecutorConfig{
		ScriptsPath:   config.Config.Scripts.Path,
		Timeout:       5 * time.Minute,
		MaxConcurrent: 5,
		DefaultEnv:    commitEnv,
	})

	// 立即返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"消息": "脚本开始执行",
		"状态": "running",
	})

	// 创建脚本执行的日志上下文
	scriptExecLogger := logger.WithFields(logrus.Fields{
		"脚本类型": config.Config.Scripts.Push,
		"提交ID": shortCommitID,
		"提交信息": pushEvent.HeadCommit.Message,
	})

	scriptExecLogger.Info("开始异步执行部署脚本")

	// 异步执行脚本
	go func() {
		result, err := executor.Execute(config.Config.Scripts.Push, "")
		if err != nil {
			scriptExecLogger.WithError(err).Error("执行脚本失败")
			return
		}

		if result.ExitCode != 0 {
			scriptExecLogger.WithFields(logrus.Fields{
				"退出码":  result.ExitCode,
				"错误信息": result.Error,
			}).Error("脚本执行返回非零退出码")
			return
		}

		scriptExecLogger.WithField("日志行数", len(result.Logs)).Info("脚本执行成功完成")
	}()
}
