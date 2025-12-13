package gomappergen

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

func MakeJenFile(currentPkg *packages.Package, config Config) *jen.File {
	pkgName := strings.ReplaceAll(config.Output.PkgName, Placeholder.CurrentPackageName, currentPkg.Name)

	return jen.NewFilePathName(currentPkg.PkgPath, pkgName)
}

func Generate(file *jen.File, currentPkg *packages.Package, config Config) error {
	fmt.Println("Generating...")
	//for _, cf := range config.Structs {
	//	targetStruct := parse.Struct(currentPkg, cf.TargetStructName)
	//	sourceStruct := parse.Struct(currentPkg, cf.SourceStructName)
	//
	//	//targetStructType := parse.StructType(currentPkg, cf.TargetStructName)
	//	//sourceStructType := parse.StructType(currentPkg, cf.SourceStructName)
	//	targetFields := parse.StructFields(currentPkg, targetStruct)
	//	sourceFields := parse.StructFields(currentPkg, sourceStruct)
	//
	//	fmt.Println(targetFields)
	//	fmt.Println(sourceFields)
	//}
	return nil
}

func mapFields(targetFields, sourceFields map[string]types.Type, match NameMatch) map[string]string {
	fmt.Println(targetFields)
	fmt.Println(sourceFields)
	return nil
}
