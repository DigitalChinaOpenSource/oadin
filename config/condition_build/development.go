//go:build !(local || production)

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println("Launch Development Environment")
	return "config-dev.yaml"
}
