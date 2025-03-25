package main

import (
	cli "aipc/byze/cmd/cli/core"
	"os"
)

func main() {
	command := cli.NewCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
