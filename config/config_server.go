package config

var ConfigRootInstance *ConfigRoot

type ConfigRoot struct {
	Vega VegaConfig `yaml:"vega" mapstructure:"vega"`
	Oss  OssConfig  `yaml:"oss" mapstructure:"oss"`
}

type VegaConfig struct {
	Url string `yaml:"url"`
}

type OssConfig struct {
	Endpoint string `yaml:"endpoint"`
}
