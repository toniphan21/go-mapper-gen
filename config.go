//go:generate rm -rf pkg/pkl
//go:generate pkl-gen-go pkl/Config.pkl
package gomappergen

import (
	"context"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/toniphan21/go-mapper-gen/pkg/pkl"
	"github.com/toniphan21/go-mapper-gen/pkg/pkl/mapper"
)

type Config struct {
	BuiltInConverters   BuiltInConverterConfig
	LibraryConverters   LibraryConverterConfig
	ConverterFunctions  []ConvertFunctionConfig
	ConverterPriorities []string
	Packages            map[string][]PackageConfig
}

type Output struct {
	PkgName      string
	FileName     string
	TestFileName string
}

type BuiltInConverterConfig struct {
	UseIdentical     bool
	UseSlice         bool
	UseTypeToPointer bool
	UsePointerToType bool
	UseNumeric       bool
	UseFunctions     bool
}

type LibraryConverterConfig struct {
	UseGRPC   bool
	UsePGType bool
	UseSQL    bool
}

func (c *BuiltInConverterConfig) EnableAll() {
	c.UseIdentical = true
	c.UseSlice = true
	c.UseTypeToPointer = true
	c.UsePointerToType = true
	c.UseNumeric = true
	c.UseFunctions = true
}

type PackageConfig struct {
	Mode                   Mode
	Output                 Output
	InterfaceName          string
	ImplementationName     string
	ConstructorName        string
	DecoratorMode          DecoratorMode
	DecoratorInterfaceName string
	DecoratorNoOpName      string
	Structs                []StructConfig
	GenerateGoDoc          bool
}

type FieldConfig struct {
	NameMatch NameMatch
	ManualMap map[string]string
}

func (c FieldConfig) Flip() FieldConfig {
	if c.ManualMap == nil {
		return FieldConfig{NameMatch: c.NameMatch, ManualMap: nil}
	}
	mm := make(map[string]string)
	for k, v := range c.ManualMap {
		mm[v] = k
	}
	return FieldConfig{NameMatch: c.NameMatch, ManualMap: mm}
}

type ConvertFunctionConfig struct {
	PackagePath string
	TypeName    string
}

type StructConfig struct {
	MapperName       string
	TargetPkgPath    string
	TargetStructName string
	SourcePkgPath    string
	SourceStructName string

	SourceToTargetFuncName   string
	SourceFromTargetFuncName string
	DecorateFuncName         string

	Pointer Pointer
	Fields  FieldConfig

	SourceFieldInterceptors map[string]FieldInterceptor
	TargetFieldInterceptors map[string]FieldInterceptor

	UseGetter bool

	GenerateSourceToTarget   bool
	GenerateSourceFromTarget bool
}

type Mode int

const (
	ModeTypes Mode = iota
	ModeFunctions
)

type DecoratorMode int

const (
	DecoratorModeAdaptive DecoratorMode = iota
	DecoratorModeNever
	DecoratorModeAlways
)

type Pointer int

const (
	PointerNone Pointer = iota
	PointerSourceOnly
	PointerTargetOnly
	PointerBoth
)

type NameMatch int

const (
	NameMatchIgnoreCase NameMatch = iota
	NameMatchExact
)

type defaultCfValue struct {
	Output                   Output
	Mode                     Mode
	InterfaceName            string
	ImplementationName       string
	ConstructorName          string
	SourceToTargetFuncName   string
	SourceFromTargetFuncName string
	DecoratorMode            DecoratorMode
	DecoratorInterfaceName   string
	DecoratorNoOpName        string
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
	Mode:                     ModeTypes,
	InterfaceName:            "iMapper",
	ImplementationName:       "iMapperImpl",
	ConstructorName:          "new_iMapper",
	SourceToTargetFuncName:   "To{TargetStructName}",
	SourceFromTargetFuncName: "From{TargetStructName}",
	DecoratorMode:            DecoratorModeAdaptive,
	DecoratorInterfaceName:   "iMapperDecorator",
	DecoratorNoOpName:        "iMapperDecoratorNoOp",
	DecorateFuncName:         "decorate{FunctionName}",
	TargetPkgPath:            Placeholder.CurrentPackage,
}

func ParseConfig(path string, provider FieldInterceptorProvider) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	cfg, err := pkl.LoadFromPath(context.Background(), path)
	if err != nil {
		return nil, err
	}
	return MakeConfig(cfg, provider)
}

func MakeConfig(cfg pkl.Config, provider FieldInterceptorProvider) (*Config, error) {
	m := &configMapper{
		provider: provider,
	}
	pkgConfigs, err := m.mapPackagesConfig(cfg.Packages, cfg.All)
	if err != nil {
		return nil, err
	}

	return &Config{
		BuiltInConverters:   m.mapBuiltInConverterConfig(cfg.Converter.BuiltIn),
		LibraryConverters:   m.mapLibraryConverterConfig(cfg.Converter.BuiltIn),
		ConverterFunctions:  m.mapConverterFunctions(cfg.Converter.Functions),
		ConverterPriorities: cfg.Converter.Priorities,
		Packages:            pkgConfigs,
	}, nil
}

type configMapper struct {
	provider FieldInterceptorProvider
}

func (m *configMapper) mapBuiltInConverterConfig(in mapper.BuiltInConverter) BuiltInConverterConfig {
	return BuiltInConverterConfig{
		UseIdentical:     in.EnableIdentical,
		UseSlice:         in.EnableSlice,
		UseTypeToPointer: in.EnableTypeToPointer,
		UsePointerToType: in.EnablePointerToType,
		UseNumeric:       in.EnableNumeric,
		UseFunctions:     in.EnableFunctions,
	}
}

func (m *configMapper) mapLibraryConverterConfig(in mapper.BuiltInConverter) LibraryConverterConfig {
	return LibraryConverterConfig{
		UseGRPC:   in.Library.EnableGrpc,
		UsePGType: in.Library.EnablePgtype,
		UseSQL:    in.Library.EnableSql,
	}
}

func (m *configMapper) mapConverterFunctions(list *[]string) []ConvertFunctionConfig {
	if list == nil {
		return nil
	}

	var result []ConvertFunctionConfig
	for _, v := range *list {
		result = append(result, parseConverterFunctionConfigFromString(v))
	}
	return result
}

func (m *configMapper) mapPackagesConfig(packages map[string]pkl.Package, all pkl.All) (map[string][]PackageConfig, error) {
	var result = make(map[string][]PackageConfig)
	if packages == nil {
		return result, nil
	}

	var pkgCfs []PackageConfig

	for path, pkg := range packages {
		if pkg.GetPriorities() != nil {
			priorities := pkg.GetPriorities()
			var priorityKeys []int
			for i := range priorities {
				priorityKeys = append(priorityKeys, i)
			}
			sort.Ints(priorityKeys)

			for _, i := range priorityKeys {
				ci, ok := priorities[i].(pkl.Package)
				if !ok {
					continue
				}

				cc := m.mapMapper(ci, all)
				if cc != nil {
					pkgCfs = append(pkgCfs, *cc)
				}
			}
		}

		defaultCf := m.mapMapper(pkg, all)
		if defaultCf == nil {
			continue
		}
		pkgCfs = append(pkgCfs, *defaultCf)
		result[path] = pkgCfs
	}
	return result, nil
}

func (m *configMapper) mapMapper(cf pkl.Package, all pkl.All) *PackageConfig {
	if len(cf.GetStructs()) == 0 {
		return nil
	}

	pkgCf := PackageConfig{
		Output:                 m.mergeOutput(&all.Output, cf.GetOutput()),
		Mode:                   m.mapMode(cf.GetMode()),
		InterfaceName:          cf.GetInterfaceName(),
		ImplementationName:     cf.GetImplementationName(),
		ConstructorName:        cf.GetConstructorName(),
		DecoratorMode:          m.mapDecoratorMode(cf.GetDecoratorMode()),
		DecoratorInterfaceName: cf.GetDecoratorInterfaceName(),
		DecoratorNoOpName:      cf.GetDecoratorNoopName(),
		GenerateGoDoc:          cf.GetGenerateGoDoc(),
	}

	var structs []StructConfig
	mapperMap := cf.GetStructs()
	mapperNames := slices.Collect(maps.Keys(mapperMap))
	sort.Strings(mapperNames)

	for _, mapperName := range mapperNames {
		v := mapperMap[mapperName]
		targetStructName := mapperName
		if v.TargetStructName != nil {
			targetStructName = *v.TargetStructName
		}
		structCf := StructConfig{
			MapperName:               mapperName,
			TargetPkgPath:            mergeConfigValue(v.TargetPkg, cf.GetTargetPkg()),
			TargetStructName:         targetStructName,
			SourcePkgPath:            mergeConfigValue(v.SourcePkg, cf.GetSourcePkg()),
			SourceStructName:         mergeConfigValue(v.SourceStructName, targetStructName),
			SourceToTargetFuncName:   mergeConfigValue(v.SourceToTargetFunctionName, cf.GetSourceToTargetFunctionName()),
			SourceFromTargetFuncName: mergeConfigValue(v.SourceFromTargetFunctionName, cf.GetSourceFromTargetFunctionName()),
			DecorateFuncName:         mergeConfigValue(v.DecorateFunctionName, cf.GetDecorateFunctionName()),
			Pointer:                  m.mapPointer(v.Pointer),
			Fields:                   m.mapFieldConfig(v.Fields),
			SourceFieldInterceptors:  m.mergeFieldInterceptor(v.SourceFields, v.Fields.Source),
			TargetFieldInterceptors:  m.mergeFieldInterceptor(v.TargetFields, v.Fields.Target),
			UseGetter:                mergeConfigValue(v.UseGetterIfAvailable, cf.GetUseGetterIfAvailable()),
			GenerateSourceToTarget:   mergeConfigValue(v.GenerateSourceToTarget, cf.GetGenerateSourceToTarget()),
			GenerateSourceFromTarget: mergeConfigValue(v.GenerateSourceFromTarget, cf.GetGenerateSourceFromTarget()),
		}

		structs = append(structs, structCf)
	}

	pkgCf.Structs = structs
	return &pkgCf
}

func (m *configMapper) mapMode(val string) Mode {
	switch val {
	case "types":
		return ModeTypes
	case "functions":
		return ModeFunctions
	default:
		return ModeTypes
	}
}

func (m *configMapper) mapDecoratorMode(val string) DecoratorMode {
	switch val {
	case "adaptive":
		return DecoratorModeAdaptive
	case "never":
		return DecoratorModeNever
	case "always":
		return DecoratorModeAlways
	default:
		return DecoratorModeAdaptive
	}
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

func (m *configMapper) mapNameMatch(val string) NameMatch {
	switch val {
	case "ignore-case":
		return NameMatchIgnoreCase
	case "exact":
		return NameMatchExact
	default:
		return NameMatchIgnoreCase
	}
}

func (m *configMapper) mergeOutput(outputs ...*pkl.Output) Output {
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

func (m *configMapper) mapFieldConfig(in mapper.Fields) FieldConfig {
	var manualMap map[string]string
	if in.Map != nil {
		manualMap = make(map[string]string)
		for k, v := range *in.Map {
			manualMap[k] = v
		}
	}
	return FieldConfig{
		NameMatch: m.mapNameMatch(in.Match),
		ManualMap: manualMap,
	}
}

func (m *configMapper) mergeFieldInterceptor(inputs ...*map[string]mapper.FieldInterceptor) map[string]FieldInterceptor {
	var result = make(map[string]FieldInterceptor)
	for _, v := range inputs {
		mapped := m.mapFieldInterceptor(v)
		for k, i := range mapped {
			result[k] = i
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (m *configMapper) mapFieldInterceptor(in *map[string]mapper.FieldInterceptor) map[string]FieldInterceptor {
	if in == nil {
		return nil
	}

	input := *in
	if input == nil {
		return nil
	}

	var result = make(map[string]FieldInterceptor)
	for k, v := range input {
		i := m.provider.MakeFieldInterceptor(v.Type, v.Options)
		if i != nil {
			result[k] = i
		}
	}
	return result
}

func mergeConfigValue[T any](structLevelValue *T, pkgLevelValue T) T {
	if structLevelValue == nil {
		return pkgLevelValue
	}
	return *structLevelValue
}

func parseConverterFunctionConfigFromString(input string) ConvertFunctionConfig {
	result := ConvertFunctionConfig{}

	s := input

	lastSlash := strings.LastIndex(s, "/")
	separatorIndex := strings.LastIndex(s, ".")

	if separatorIndex > lastSlash {
		result.PackagePath = s[:separatorIndex]
		s = s[separatorIndex+1:]
	} else {
		result.PackagePath = ""
	}

	result.TypeName = s
	return result
}
