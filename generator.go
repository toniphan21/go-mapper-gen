package gomappergen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

func Generate(file *jen.File, currentPkg *packages.Package, config []Config) error {
	fmt.Println("Generating...")
	return nil
}
