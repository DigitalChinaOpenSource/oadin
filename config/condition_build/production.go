//go:build production

package condition_build

import "fmt"

func GetConfigFile() string {
	fmt.Println("Launch Production Environment")
	return "config-prod.yaml"
}
