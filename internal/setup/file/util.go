package file

import (
	"os"
	"path/filepath"
)

func ContentFromTestData(elem ...string) []byte {
	elems := []string{"testdata"}
	elems = append(elems, elem...)
	c, err := os.ReadFile(filepath.Join(elems...))
	if err != nil {
		panic(err)
	}
	return c
}
