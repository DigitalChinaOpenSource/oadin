//go:build !local && !development

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println(" setup production config file")
	return "config-prod.yaml"
}
