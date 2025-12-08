package pkl

import (
	"embed"
	_ "embed"
)

//go:embed *
var FS embed.FS

func LibFilePaths() []string {
	return []string{
		"Config.pkl",
	}
}
