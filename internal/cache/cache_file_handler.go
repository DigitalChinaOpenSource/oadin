package cache

import (
	"encoding/json"
	"oadin/config"
	"os"
	"path/filepath"
)

const (
	SYSTEM_DIRECTORY = "system"
	USER_DIRECTORY   = "user"
	SETTINGS_FILE    = "settings.json"
)

func CacheFilePath() string {
	return filepath.Join(config.GlobalOadinEnvironment.RootDir, "cache")
}

// 获取本地配置文件路径
func getConfigFilePath(appName string, filename string) (string, error) {
	configDir := CacheFilePath()

	dir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// SaveAsJSON 写入任意结构体为JSON到本地
func SaveAsJSON(data interface{}, appName string, filename string) error {
	path, err := getConfigFilePath(appName, filename)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 格式化输出
	return encoder.Encode(data)
}

// LoadFromJSON 读取JSON文件填充到结构体
func LoadFromJSON(data interface{}, appName string, filename string) error {
	path, err := getConfigFilePath(appName, filename)
	if err != nil {
		return err
	}

	// 检查文件是否存在
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// 文件不存在，创建文件并写入初始化的数据
		return SaveAsJSON(data, appName, filename)
	}

	// 文件存在，打开并读取
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 检查文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// 如果文件为空，不进行解码
	if fileInfo.Size() == 0 {
		return nil
	}

	// 文件有内容，解码
	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}

func WriteSystemSettings(data SystemSettings) error {
	return SaveAsJSON(data, SYSTEM_DIRECTORY, SETTINGS_FILE)
}

func ReadSystemSettings(data *SystemSettings) error {
	return LoadFromJSON(data, SYSTEM_DIRECTORY, SETTINGS_FILE)
}

type SystemSettings struct {
	OllamaRegistry string `json:"ollamaRegistry"`
	SystemProxy    struct {
		Enabled  bool   `json:"enabled"`
		Endpoint string `json:"endpoint"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"systemProxy"`
}
