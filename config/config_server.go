package config

var ConfigRootInstance *ConfigRoot

type ConfigRoot struct {
	Vega     VegaConfig     `yaml:"vega" mapstructure:"vega"`
	Oss      OssConfig      `yaml:"oss" mapstructure:"oss"`
	Registry RegistryConfig `yaml:"registry" mapstructure:"registry"`
	Ollama   OllamaRegistry `yaml:"ollama" mapstructure:"ollama"`
}

type VegaConfig struct {
	Url string `yaml:"url"`
}

type OssConfig struct {
	Endpoint string `yaml:"endpoint"`
	Icon     string `yaml:"icon"` // 用于获取头像等静态资源的前缀
}

type OllamaRegistry struct {
	Url string `yaml:"url" mapstructure:"url"`
}

type RegistryConfig struct {
	Npm string `yaml:"npm" mapstructure:"npm"`
	Pip string `yaml:"pip" mapstructure:"pip"`
}
