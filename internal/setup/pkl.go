package setup

import (
	"github.com/toniphan21/go-mapper-gen/internal/setup/file"
	"github.com/toniphan21/go-mapper-gen/pkl"
)

func PklLibFiles() []file.File {
	ps := pkl.LibFilePaths()
	result := make([]file.File, len(ps))
	for i, path := range ps {
		b, err := pkl.FS.ReadFile(path)
		if err != nil {
			panic(err)
		}
		result[i] = &file.Pkl{Path: path, Content: string(b)}
	}
	return result
}
