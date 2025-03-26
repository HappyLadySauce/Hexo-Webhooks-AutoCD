package config

import (
	"log"

	"github.com/spf13/viper"
)

type config struct {
	Webhook struct {
		Port   int    `mapstructure:"port"`
		Path   string `mapstructure:"path"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"webhook"`

	Scripts struct {
		Path    string `mapstructure:"path"`
		Push    string `mapstructure:"push"`
	} `mapstructure:"scripts"`

	SSL struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"ssl"`
}

var Config *config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var config config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	log.Printf("配置文件解析成功: %+v", config)
	Config = &config
}
