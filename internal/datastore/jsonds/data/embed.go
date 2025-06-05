package jsondsTemplate

import (
	"embed"
)

//go:embed *.json
var JsonDataStoreFS embed.FS
