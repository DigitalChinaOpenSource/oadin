package trayTemplate

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed *.png
var TrayIconFS embed.FS

func DebugListFiles() {
	fmt.Println("Embedded icon files:")
	fs.WalkDir(TrayIconFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fmt.Printf("  - %s\n", path)
		}
		return nil
	})
}
