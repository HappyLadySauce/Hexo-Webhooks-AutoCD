package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type config struct {
	Webhook struct {
		Port   int    `mapstructure:"port"`
		Path   string `mapstructure:"path"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"webhook"`

	Scripts struct {
		Path          string `mapstructure:"path"`
		Push          string `mapstructure:"push"`
		Timeout       string `mapstructure:"timeout"`
		MaxConcurrent int    `mapstructure:"max_concurrent"`
	} `mapstructure:"scripts"`

	Logs struct {
		Path       string `mapstructure:"path"`
		Level      string `mapstructure:"level"`
		Format     string `mapstructure:"format"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
	} `mapstructure:"logs"`

	SSL struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"ssl"`
}

var Config *config

// InitConfig 初始化配置
// 注意：因为日志系统依赖于配置，所以在配置加载时我们还不能使用日志系统
// 因此这里使用标准库的日志包作为临时解决方案
func InitConfig() {
	// 设置标准库日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("开始加载配置...")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("致命错误: 读取配置文件失败: %v\n", err)
		os.Exit(1)
	}

	var config config

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("致命错误: 解析配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 检查必要的配置项
	if config.Webhook.Port == 0 {
		log.Println("警告: Webhook端口未设置，使用默认端口8080")
		config.Webhook.Port = 8080
	}

	if config.Webhook.Path == "" {
		log.Println("警告: Webhook路径未设置，使用默认路径/webhook")
		config.Webhook.Path = "/webhook"
	}

	if config.Logs.Path == "" {
		log.Println("警告: 日志路径未设置，使用默认路径./logs/webhooks.log")
		config.Logs.Path = "./logs/webhooks.log"
	}

	if config.Logs.Level == "" {
		log.Println("警告: 日志级别未设置，使用默认级别info")
		config.Logs.Level = "info"
	}

	// 设置默认值
	if config.Logs.MaxSize == 0 {
		config.Logs.MaxSize = 100 // 默认100MB
	}

	if config.Logs.MaxBackups == 0 {
		config.Logs.MaxBackups = 5 // 默认保留5个备份
	}

	if config.Logs.MaxAge == 0 {
		config.Logs.MaxAge = 30 // 默认保留30天
	}

	log.Println("配置文件加载成功")
	Config = &config
}
