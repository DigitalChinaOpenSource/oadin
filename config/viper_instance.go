package config

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed *
var configFs embed.FS

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

	//使用 viper
	ViperInstance := viper.New()
	ViperInstance.SetConfigType("yaml")

	file, err := configFs.Open(configFile)
	if err != nil {
		panic(fmt.Errorf("open config file failed: %s \n", err))
	}

	if err := ViperInstance.ReadConfig(file); err != nil {
		panic(fmt.Errorf("read config failed: %s \n", err))
	}

	ConfigRootInstance = new(ConfigRoot)
	if err := ViperInstance.Unmarshal(&ConfigRootInstance); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %s \n", err))
	}

}

// 获取项目根目录（跨平台）
func getProjectRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)              // 获取当前文件路径
	projectRoot := filepath.Dir(filepath.Dir(currentFile)) // 上溯到项目根目录
	return projectRoot
}
