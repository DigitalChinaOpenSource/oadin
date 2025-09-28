//go:build local

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println("Launch Local Environment")
	return "config-local.yaml"
}
