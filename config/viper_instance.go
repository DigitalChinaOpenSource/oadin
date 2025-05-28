package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

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
		configFile = "config-dev.yaml"
	}

	// 2. 动态获取配置文件路径（相对于项目根目录）
	configDir, err := filepath.Abs("./config") // 转为绝对路径（适用于 Windows/Linux）
	if err != nil {
		panic(fmt.Errorf("无法解析 config 目录: %v", err))
	}
	//使用 viper
	ViperInstance := viper.New()
	ViperInstance.AddConfigPath("./config")
	ViperInstance.SetConfigFile(filepath.Join(configDir, configFile))
	ViperInstance.SetConfigType("yaml")
	if err := ViperInstance.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}
	ConfigRootInstance = new(ConfigRoot)
	if err := ViperInstance.Unmarshal(&ConfigRootInstance); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %s \n", err))
	}

}
