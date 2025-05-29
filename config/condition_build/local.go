//go:build local

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println(" setup local config file")
	return "config-local.yaml"
}
