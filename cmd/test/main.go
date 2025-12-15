package main

import (
	"fmt"
	"go/importer"
)

func main() {
	imp := importer.Default()
	pkg, err := imp.Import("github.com/google/uuid")
	fmt.Println(err)
	fmt.Println(pkg.Scope().Names())
}
