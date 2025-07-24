package config

var ConfigRootInstance *ConfigRoot

type ConfigRoot struct {
	Vega     VegaConfig     `yaml:"vega" mapstructure:"vega"`
	Oss      OssConfig      `yaml:"oss" mapstructure:"oss"`
	Registry RegistryConfig `yaml:"registry" mapstructure:"registry"`
}

type VegaConfig struct {
	Url string `yaml:"url"`
}

type OssConfig struct {
	Endpoint string `yaml:"endpoint"`
	Icon     string `yaml:"icon"` // 用于获取头像等静态资源的前缀
}

type RegistryConfig struct {
	Npm string `yaml:"npm" mapstructure:"npm"`
	Pip string `yaml:"pip" mapstructure:"pip"`
}
