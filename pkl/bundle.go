package pkl

import (
	"embed"
	_ "embed"
)

//go:embed *
var FS embed.FS

func AmendsPath() string {
	return "https://github.com/toniphan21/go-mapper-gen/releases/download/v0.1.0/Config.pkl"
}

func LibFilePaths() []string {
	return []string{
		"Config.pkl",
	}
}
