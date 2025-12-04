package trayTemplate

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed *.png *.ico
var TrayIconFS embed.FS

// DebugListFiles 列出所有嵌入的文件（调试用）
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
