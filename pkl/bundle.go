package pkl

import (
	"embed"
	_ "embed"
)

//go:embed *
var FS embed.FS

func AmendsPath() string {
	return "https://github.com/toniphan21/go-mapper-gen/releases/download/current/Config.pkl"
}

func LibFilePaths() []string {
	return []string{
		"mapper.pkl",
		"Config.pkl",
	}
}
