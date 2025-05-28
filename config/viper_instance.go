package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func init() {
	var configFile string
	// 根据环境变量选择配置文件
	configEnv := os.Getenv("GO_ENV")
	switch configEnv {
	case "local":
		configFile = "./config/config-local.yaml"
	case "dev":
		configFile = "./config/config-dev.yaml"
	case "prod":
		configFile = "./config/config-prod.yaml"
	default:
		configFile = "./config/config-prod.yaml"
	}

	//使用 viper
	ViperInstance := viper.New()
	ViperInstance.SetConfigFile(configFile)
	ViperInstance.SetConfigType("yaml")
	if err := ViperInstance.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}
	ConfigRootInstance = new(ConfigRoot)
	if err := ViperInstance.Unmarshal(&ConfigRootInstance); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %s \n", err))
	}

}
