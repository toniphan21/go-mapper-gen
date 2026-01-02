package pkl

import (
	"embed"
	_ "embed"
)

//go:embed *
var FS embed.FS

func AmendsPath(base string) string {
	return base + "/Config.pkl"
}

func ImportPaths(base string) []string {
	return []string{
		base + "/set.pkl",
	}
}

func LibFilePaths() []string {
	return []string{
		"set.pkl",
		"mapper.pkl",
		"Config.pkl",
	}
}
