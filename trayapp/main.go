package main

import (
	"oadin/config"
	"oadin/tray"
)

func main() {
	logPath := config.GlobalOADINEnvironment.LogDir
	pidPath := logPath
	manager := tray.NewManager(false, logPath, pidPath)
	manager.Start()
}
