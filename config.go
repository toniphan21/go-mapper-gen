//go:generate rm -rf internal/config
//go:generate pkl-gen-go pkl/Config.pkl
package gomappergen

import (
	"context"
	"os"

	"github.com/apple/pkl-go/pkl"
	"github.com/toniphan21/go-mapper-gen/internal/config"
)

type Output struct {
	PkgName      string
	FileName     string
	TestFileName string
}

type Config struct {
	Output                 Output
	InterfaceName          string
	ImplementationName     string
	ConstructorName        string
	DecoratorInterfaceName string
	Structs                []Struct
	GenerateGoDoc          bool
}

type Struct struct {
	TargetPkgPath    string
	TargetStructName string
	SourcePkgPath    string
	SourceStructName string

	SourceToTargetFuncName   string
	SourceFromTargetFuncName string
	DecorateFuncName         string

	Pointer        Pointer
	FieldNameMatch FieldNameMatch

	GenerateSourceToTarget   bool
	GenerateSourceFromTarget bool
}

type Pointer int

const (
	PointerNone Pointer = iota
	PointerSourceOnly
	PointerTargetOnly
	PointerBoth
)

type FieldNameMatch int

const (
	FieldNameMatchIgnoreCase FieldNameMatch = iota
	FieldNameMatchExact
)

type defaultCfValue struct {
	Output                   Output
	InterfaceName            string
	ImplementationName       string
	ConstructorName          string
	SourceToTargetFuncName   string
	SourceFromTargetFuncName string
	DecoratorInterfaceName   string
	DecorateFuncName         string
	TargetPkgPath            string
}

type cfPlaceHolder struct {
	TargetStructName   string
	SourceStructName   string
	CurrentPackage     string
	CurrentPackageName string
	FunctionName       string
}

var Placeholder = cfPlaceHolder{
	TargetStructName:   "{TargetStructName}",
	SourceStructName:   "{SourceStructName}",
	CurrentPackage:     "{CurrentPackage}",
	CurrentPackageName: "{CurrentPackageName}",
	FunctionName:       "{FunctionName}",
}

var Default = defaultCfValue{
	Output: Output{
		PkgName:      Placeholder.CurrentPackageName,
		FileName:     "gen_mapper.go",
		TestFileName: "gen_mapper_test.go",
	},
	InterfaceName:            "iMapper",
	ImplementationName:       "iMapperImpl",
	ConstructorName:          "new_iMapper",
	SourceToTargetFuncName:   "To{TargetStructName}",
	SourceFromTargetFuncName: "From{TargetStructName}",
	DecoratorInterfaceName:   "iMapperDecorator",
	DecorateFuncName:         "decorate{FunctionName}",
	TargetPkgPath:            Placeholder.CurrentPackage,
}

func ParseConfig(path string) (map[string]Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	evaluator, err := pkl.NewEvaluator(context.Background(), pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(context.Background(), evaluator, pkl.FileSource(path))
	if err != nil {
		return nil, err
	}
	return (&configMapper{}).mapPackagesConfig(cfg.Packages, cfg.All)
}

type configMapper struct{}

func (m *configMapper) mapPackagesConfig(packages map[string]config.Mapper, all config.Base) (map[string]Config, error) {
	var result = make(map[string]Config)
	if packages == nil {
		return result, nil
	}

	for path, pkg := range packages {
		if len(pkg.GetStructs()) == 0 {
			continue
		}

		pkgCf := Config{
			Output:                 m.mergeOutput(all.GetOutput(), pkg.GetOutput()),
			InterfaceName:          pkg.GetInterfaceName(),
			ImplementationName:     pkg.GetImplementationName(),
			ConstructorName:        pkg.GetConstructorName(),
			DecoratorInterfaceName: pkg.GetDecoratorInterfaceName(),
			GenerateGoDoc:          pkg.GetGenerateGoDoc(),
		}

		var structs []Struct
		for targetStructName, v := range pkg.GetStructs() {
			structCf := Struct{
				TargetPkgPath:            m.mergeValue(v.TargetPkg, pkg.GetTargetPkg()),
				TargetStructName:         targetStructName,
				SourcePkgPath:            m.mergeValue(v.SourcePkg, pkg.GetSourcePkg()),
				SourceStructName:         m.mergeValue(v.SourceStructName, targetStructName),
				SourceToTargetFuncName:   m.mergeValue(v.SourceToTargetFunctionName, pkg.GetSourceToTargetFunctionName()),
				SourceFromTargetFuncName: m.mergeValue(v.SourceFromTargetFunctionName, pkg.GetSourceFromTargetFunctionName()),
				DecorateFuncName:         m.mergeValue(v.DecorateFunctionName, pkg.GetDecorateFunctionName()),
				Pointer:                  m.mapPointer(v.Pointer),
				FieldNameMatch:           m.mapFieldNameMatch(v.FieldNameMatch),
				GenerateSourceToTarget:   v.GenerateSourceToTarget,
				GenerateSourceFromTarget: v.GenerateSourceFromTarget,
			}

			structs = append(structs, structCf)
		}

		pkgCf.Structs = structs
		result[path] = pkgCf
	}
	return result, nil
}

func (m *configMapper) mapPointer(val string) Pointer {
	switch val {
	case "none":
		return PointerNone
	case "source-only":
		return PointerSourceOnly
	case "target-only":
		return PointerTargetOnly
	case "both":
		return PointerBoth
	default:
		return PointerNone
	}
}

func (m *configMapper) mapFieldNameMatch(val string) FieldNameMatch {
	switch val {
	case "ignore-case":
		return FieldNameMatchIgnoreCase
	case "exact":
		return FieldNameMatchExact
	default:
		return FieldNameMatchIgnoreCase
	}
}

func (m *configMapper) mergeValue(structLevelValue *string, pkgLevelValue string) string {
	if structLevelValue == nil {
		return pkgLevelValue
	}
	return *structLevelValue
}

func (m *configMapper) mergeOutput(outputs ...*config.Output) Output {
	output := &Output{}
	for _, o := range outputs {
		if o == nil {
			continue
		}

		if o.Package != nil {
			output.PkgName = *o.Package
		}

		if o.FileName != nil {
			output.FileName = *o.FileName
		}

		if o.TestFileName != nil {
			output.TestFileName = *o.TestFileName
		}
	}
	return *output
}
