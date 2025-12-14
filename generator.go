package gomappergen

import (
	"go/types"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/toniphan21/go-mapper-gen/internal/parse"
	"golang.org/x/tools/go/packages"
)

func Generate(currentPkg *packages.Package, configs []Config, fileManager FileManager) error {
	for _, cf := range configs {
		file := fileManager.MakeJenFile(currentPkg, cf)
		if file == nil {
			continue
		}

		if err := generateMapper(file, currentPkg, cf); err != nil {
			return err
		}
	}
	return nil
}

type convertibleField struct {
	targetSymbol Symbol
	sourceSymbol Symbol
	converter    Converter
}

type genMapFunc struct {
	name               string
	funcName           string
	decorateFuncName   string
	targetParamName    string
	targetType         types.Type
	targetPointer      bool
	sourceParamName    string
	sourceType         types.Type
	sourcePointer      bool
	mappedFields       []convertibleField
	missingFields      []string
	unconvertibleField []string
}

func generateMapper(file *jen.File, currentPkg *packages.Package, config Config) error {
	mapFuncs, err := collectMapFuncs(currentPkg, config)
	if err != nil {
		return err
	}

	if len(mapFuncs) == 0 {
		return nil
	}

	slices.SortFunc(mapFuncs, func(a, b genMapFunc) int {
		return strings.Compare(a.name, b.name)
	})

	generateMapperInterface(file, config.InterfaceName, mapFuncs)

	return nil
}

func generateMapperInterface(file *jen.File, name string, mapFuncs []genMapFunc) {
	var signatures = jen.Line()

	for _, mf := range mapFuncs {
		var p, r []jen.Code

		if mf.sourcePointer {
			p = append(p, jen.Id(mf.sourceParamName).Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceType))))
		} else {
			p = append(p, jen.Id(mf.sourceParamName).Add(GeneratorUtil.TypeToJenCode(mf.sourceType)))
		}

		if mf.targetPointer {
			r = append(r, jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetType)))
		} else {
			r = append(r, jen.Add(GeneratorUtil.TypeToJenCode(mf.targetType)))
		}
		signatures = signatures.Id(mf.funcName).Params(p...).Params(r...).Line().Line()
	}

	file.Type().Id(name).Interface(signatures).Line()
}

func collectMapFuncs(currentPkg *packages.Package, config Config) ([]genMapFunc, error) {
	var mapFuncs []genMapFunc
	for _, cf := range config.Structs {
		var vars = map[string]string{
			Placeholder.CurrentPackage:     currentPkg.PkgPath,
			Placeholder.CurrentPackageName: currentPkg.Name,
			Placeholder.TargetStructName:   cf.TargetStructName,
			Placeholder.SourceStructName:   cf.SourceStructName,
		}

		var useTargetPointer, useSourcePointer bool
		switch cf.Pointer {
		case PointerSourceOnly:
			useSourcePointer = true
		case PointerTargetOnly:
			useTargetPointer = true
		case PointerBoth:
			useTargetPointer = true
			useSourcePointer = true
		default:
			useTargetPointer = false
			useSourcePointer = false
		}

		targetStruct := parse.Struct(currentPkg, cf.TargetStructName)
		sourceStruct := parse.Struct(currentPkg, cf.SourceStructName)

		targetStructType := parse.StructType(currentPkg, cf.TargetStructName)
		sourceStructType := parse.StructType(currentPkg, cf.SourceStructName)
		targetFields := parse.StructFields(currentPkg, targetStruct)
		sourceFields := parse.StructFields(currentPkg, sourceStruct)

		if cf.GenerateSourceToTarget {
			toTargetFuncName := replacePlaceholders(cf.SourceToTargetFuncName, vars)

			tv := vars
			tv[Placeholder.FunctionName] = toTargetFuncName
			decorateToTargetFuncName := replacePlaceholders(cf.DecorateFuncName, tv)

			mapFunc := genMapFunc{
				name:             cf.MapperName + "-SourceToTarget",
				funcName:         toTargetFuncName,
				decorateFuncName: decorateToTargetFuncName,
				targetParamName:  "out",
				targetType:       targetStructType,
				targetPointer:    useTargetPointer,
				sourceParamName:  "in",
				sourceType:       sourceStructType,
				sourcePointer:    useSourcePointer,
			}

			fillMapFunc(&mapFunc, targetFields, sourceFields, cf.Fields)
			mapFuncs = append(mapFuncs, mapFunc)
		}

		if cf.GenerateSourceFromTarget {
			fromTargetFuncName := replacePlaceholders(cf.SourceFromTargetFuncName, vars)

			fv := vars
			fv[Placeholder.FunctionName] = fromTargetFuncName
			decorateFromTargetFuncName := replacePlaceholders(cf.DecorateFuncName, fv)

			mapFunc := genMapFunc{
				name:             cf.MapperName + "-TargetToSource",
				funcName:         fromTargetFuncName,
				decorateFuncName: decorateFromTargetFuncName,
				targetParamName:  "out",
				targetType:       sourceStructType,
				targetPointer:    useSourcePointer,
				sourceParamName:  "in",
				sourceType:       targetStructType,
				sourcePointer:    useTargetPointer,
			}

			fillMapFunc(&mapFunc, sourceFields, targetFields, cf.Fields.Flip())
			mapFuncs = append(mapFuncs, mapFunc)
		}
	}
	return mapFuncs, nil
}

func fillMapFunc(mapFunc *genMapFunc, targetFields, sourceFields map[string]types.Type, config FieldConfig) {
	mappedFields := mapFieldNames(targetFields, sourceFields, config)
	for target, source := range mappedFields {
		if source == "" {
			mapFunc.missingFields = append(mapFunc.missingFields, target)
			continue
		}

		tt, ok := targetFields[target]
		if !ok {
			continue
		}
		st, ok := sourceFields[target]
		if !ok {
			continue
		}

		converter, ok := FindConverter(tt, st)
		if !ok {
			mapFunc.unconvertibleField = append(mapFunc.unconvertibleField, target)
			continue
		}

		mapFunc.mappedFields = append(mapFunc.mappedFields, convertibleField{
			targetSymbol: newSymbol("out", target, tt),
			sourceSymbol: newSymbol("in", source, st),
			converter:    converter,
		})
	}
}

func mapFieldNames(targetFields, sourceFields map[string]types.Type, config FieldConfig) map[string]string {
	result := make(map[string]string)
	for target, _ := range targetFields {
		result[target] = ""
		if config.ManualMap != nil {
			manualSource, ok := config.ManualMap[target]
			if ok {
				if _, have := sourceFields[manualSource]; have {
					result[target] = manualSource
				}
				continue
			}
		}

		for source, _ := range sourceFields {
			if config.NameMatch == NameMatchIgnoreCase && strings.ToLower(target) == strings.ToLower(source) {
				result[target] = source
				break
			}

			if config.NameMatch == NameMatchExact && target == source {
				result[target] = source
				break
			}
		}
	}
	return result
}
