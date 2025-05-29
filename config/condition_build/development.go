//go:build development

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println(" setup develop config file")
	return "config-dev.yaml"
}
