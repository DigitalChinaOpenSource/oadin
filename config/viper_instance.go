package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var ViperInstance *viper.Viper

func init() {
	var configFile string
	// 根据环境变量选择配置文件
	configEnv := os.Getenv("GO_ENV")
	switch configEnv {
	case "local":
		configFile = "config-local.yaml"
	case "dev":
		configFile = "config-dev.yaml"
	case "prod":
		configFile = "config-prod.yaml"
	default:
		configFile = "config.yaml"
	}

	//使用 viper
	ViperInstance := viper.New()
	ViperInstance.SetConfigFile(configFile)
	ViperInstance.AddConfigPath("./config")
	ViperInstance.SetConfigType("yaml")
	if err := ViperInstance.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

}
