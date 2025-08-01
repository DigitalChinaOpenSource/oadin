package config

import (
	"embed"
	_ "embed"
	"fmt"

	"github.com/spf13/viper"
	"oadin/config/condition_build"
)

//go:embed *
var configFs embed.FS

func init() {
	var configFile string
	// 根据条件编译来选择配置文件
	configFile = condition_build.GetConfigFile()

	// 使用 viper
	ViperInstance := viper.New()
	ViperInstance.SetConfigType("yaml")

	file, err := configFs.Open(configFile)
	if err != nil {
		panic(fmt.Errorf("open config file failed: %s \n", err))
	}

	if err := ViperInstance.ReadConfig(file); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

	// 初始化数据到配置结构体里面
	ConfigRootInstance = new(ConfigRoot)
	if err := ViperInstance.Unmarshal(&ConfigRootInstance); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %s \n", err))
	}
}
