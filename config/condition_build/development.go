//go:build !local && !prodcution

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println(" setup develop config file")
	return "config-dev.yaml"
}
