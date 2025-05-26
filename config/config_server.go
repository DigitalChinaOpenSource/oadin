package config

type ConfigRoot struct {
	Vega VegaConfig `yaml:"vega"`
	Oss  OssConfig  `yaml:"oss"`
}

type VegaConfig struct {
	Url string `yaml:"url"`
}

type OssConfig struct {
	Endpoint string `yaml:"endpoint"`
}

func GetVegaConfig() VegaConfig {
	vegaConfig := &VegaConfig{}
	err := ViperInstance.Unmarshal(vegaConfig)
	if err != nil {
		panic("failed to unmarshal VegaConfig: " + err.Error())
	}
	return *vegaConfig
}

func GetOssConfig() OssConfig {

	ossConfig := &OssConfig{}
	err := ViperInstance.Unmarshal(ossConfig)
	if err != nil {
		panic("failed to unmarshal OssConfig: " + err.Error())
	}
	return *ossConfig
}
