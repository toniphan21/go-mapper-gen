package gomappergen

import (
	"context"
	"fmt"
	"go/types"
	"math"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

func Generate(parser Parser, fileManager FileManager, currentPkg *packages.Package, configs []PackageConfig) error {
	for _, cf := range configs {
		file := fileManager.MakeJenFile(currentPkg, cf)
		if file == nil {
			continue
		}

		if err := generateMapper(parser, file, currentPkg, cf); err != nil {
			return err
		}
	}
	return nil
}

type convertibleField struct {
	index           int
	targetFieldName string
	targetSymbol    Symbol
	sourceFieldName string
	sourceSymbol    Symbol
	converter       Converter
}

type genMapFunc struct {
	name                string
	funcName            string
	decorateFuncName    string
	targetPkgPath       string
	targetParamName     string
	targetType          types.Type
	targetPointer       bool
	sourcePkgPath       string
	sourceParamName     string
	sourceType          types.Type
	sourcePointer       bool
	mappedFields        []convertibleField
	missingFields       []string
	unconvertibleFields []string
	targetFieldsIndex   map[string]int
	sourceFieldsIndex   map[string]int
}

func (mf *genMapFunc) paramsAndResults() ([]jen.Code, []jen.Code) {
	var params, result []jen.Code

	if mf.sourcePointer {
		params = append(params, jen.Id(mf.sourceParamName).Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceType))))
	} else {
		params = append(params, jen.Id(mf.sourceParamName).Add(GeneratorUtil.TypeToJenCode(mf.sourceType)))
	}

	if mf.targetPointer {
		result = append(result, jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetType)))
	} else {
		result = append(result, jen.Add(GeneratorUtil.TypeToJenCode(mf.targetType)))
	}

	return params, result
}

func (mf *genMapFunc) appendUnconvertibleField(field string) {
	mf.unconvertibleFields = append(mf.unconvertibleFields, field)
}

func generateMapper(parser Parser, file *jen.File, currentPkg *packages.Package, config PackageConfig) error {
	ctx := &converterContext{
		Context: context.Background(),
		jenFile: file,
		parser:  parser,
		logger:  nil,
	}
	mapFuncs, err := collectMapFuncs(ctx, currentPkg, config)
	if err != nil {
		return err
	}

	if len(mapFuncs) == 0 {
		return nil
	}

	slices.SortFunc(mapFuncs, func(a, b *genMapFunc) int {
		return strings.Compare(a.name, b.name)
	})

	generateMapperInterface(file, config, mapFuncs)
	generateDecoratorInterface(ctx, config, mapFuncs)
	generateMapperConstructor(ctx, config, mapFuncs)
	generateMapperImplementation(ctx, config, mapFuncs)
	generateDecoratorNoOp(ctx, config, mapFuncs)
	generateCompileTimeCheck(file, config, mapFuncs)

	return nil
}

func generateMapperInterface(file *jen.File, config PackageConfig, mapFuncs []*genMapFunc) {
	var signatures []jen.Code

	for _, mf := range mapFuncs {
		params, results := mf.paramsAndResults()

		if config.GenerateGoDoc {
			comment := fmt.Sprintf(
				"%v converts a %v value into a %v value.",
				mf.funcName,
				GeneratorUtil.SimpleName(mf.sourceType),
				GeneratorUtil.SimpleName(mf.targetType),
			)

			signatures = append(signatures, GeneratorUtil.WrapComment(comment))
		}
		signatures = append(signatures, jen.Id(mf.funcName).Params(params...).Params(results...).Line())
	}

	file.Type().Id(config.InterfaceName).Interface(signatures...).Line().Line()
}

func generateDecoratorInterface(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc) {
	if !hasMissingOrUnconvertibleField(mapFuncs) {
		return
	}

	var signatures []jen.Code

	for _, mf := range mapFuncs {
		var params []jen.Code
		params = append(params, jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceType))))
		params = append(params, jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetType))))

		signatures = append(signatures, jen.Id(mf.decorateFuncName).Params(params...).Params().Line())
	}

	ctx.jenFile.Type().Id(config.DecoratorInterfaceName).Interface(signatures...).Line().Line()
}

func generateMapperConstructor(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc) {
	var params, body []jen.Code

	if hasMissingOrUnconvertibleField(mapFuncs) {
		params = append(params, jen.Id("decorator").Id(config.DecoratorInterfaceName))
		body = append(body, jen.Return(jen.Op("&").Id(config.ImplementationName).Values(jen.DictFunc(func(d jen.Dict) {
			d[jen.Id("decorator")] = jen.Id("decorator")
		}))))
	} else {
		body = append(body, jen.Return(jen.Op("&").Id(config.ImplementationName).Block(nil)))
	}

	ctx.jenFile.Func().Id(config.ConstructorName).Params(params...).Params(jen.Id(config.InterfaceName)).Block(body...)
}

func generateMapperImplementation(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc) {
	file := ctx.JenFile()
	useDecorator := hasMissingOrUnconvertibleField(mapFuncs)

	if useDecorator {
		file.Type().Id(config.ImplementationName).Struct(
			jen.Id("decorator").Add(jen.Id(config.DecoratorInterfaceName)),
		).Line()
	} else {
		file.Type().Id(config.ImplementationName).Struct().Line()
	}

	for _, mf := range mapFuncs {
		params, results := mf.paramsAndResults()

		var body []jen.Code
		body = append(body, jen.Var().Id(mf.targetParamName).Add(GeneratorUtil.TypeToJenCode(mf.targetType)).Line())

		for _, field := range mf.mappedFields {
			opt := ConverterOption{}

			convertedCode := field.converter.ConvertField(ctx, field.targetSymbol, field.sourceSymbol, opt)
			if convertedCode != nil {
				body = append(body, convertedCode)
			}
		}

		if len(mf.missingFields) > 0 || len(mf.unconvertibleFields) > 0 {
			body = append(body, jen.Line())
			body = append(body, jen.If(jen.Id("m").Dot("decorator").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
				g.Id("m").Dot("decorator").Dot(mf.decorateFuncName).Params(
					jen.Op("&").Id(mf.sourceParamName),
					jen.Op("&").Id(mf.targetParamName),
				)
			}))
		}

		if mf.targetPointer {
			body = append(body, jen.Line().Return(jen.Op("&").Id(mf.targetParamName)))
		} else {
			body = append(body, jen.Line().Return(jen.Id(mf.targetParamName)))
		}

		file.Func().
			Params(jen.Id("m").Op("*").Id(config.ImplementationName)).
			Id(mf.funcName).
			Params(params...).
			Params(results...).
			Block(body...).
			Line()

		ctx.resetVarCount()
	}
}

func generateDecoratorNoOp(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc) {
	if !hasMissingOrUnconvertibleField(mapFuncs) || config.DecoratorNoOpName == "" {
		return
	}

	ctx.jenFile.Type().Id(config.DecoratorNoOpName).Struct().Line()

	for _, mf := range mapFuncs {
		var body []jen.Code
		var params []jen.Code
		params = append(params, jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceType))))
		params = append(params, jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetType))))

		hasMissingField := len(mf.missingFields) > 0
		hasUnconvertibleField := len(mf.unconvertibleFields) > 0

		if hasMissingField {
			body = append(body, jen.Comment("Fields that could not be mapped:"))

			fields := sortFieldsByIndex(mf.missingFields, mf.targetFieldsIndex)
			for _, field := range fields {
				body = append(body, jen.Comment("out."+field+" = "))
			}
		}

		if hasMissingField && hasUnconvertibleField {
			body = append(body, jen.Line())
		}

		if hasUnconvertibleField {
			body = append(body, jen.Comment("Fields that could not be converted (no suitable converter found):"))

			fields := sortFieldsByIndex(mf.unconvertibleFields, mf.targetFieldsIndex)
			for _, field := range fields {
				body = append(body, jen.Comment("out."+field+" = "))
			}
		}

		ctx.JenFile().Func().
			Params(jen.Id("d").Op("*").Id(config.DecoratorNoOpName)).
			Id(mf.decorateFuncName).
			Params(params...).
			Block(body...).
			Line()

		ctx.resetVarCount()
	}
}

func generateCompileTimeCheck(file *jen.File, config PackageConfig, mapFuncs []*genMapFunc) {
	file.Var().Id("_").Id(config.InterfaceName).Op("=").Parens(jen.Op("*").Id(config.ImplementationName)).Parens(jen.Nil())

	if !hasMissingOrUnconvertibleField(mapFuncs) || config.DecoratorNoOpName == "" {
		return
	}
	file.Var().Id("_").Id(config.DecoratorInterfaceName).Op("=").Parens(jen.Op("*").Id(config.DecoratorNoOpName)).Parens(jen.Nil())
}

func collectMapFuncs(ctx ConverterContext, currentPkg *packages.Package, config PackageConfig) ([]*genMapFunc, error) {
	var mapFuncs []*genMapFunc
	for _, cf := range config.Structs {
		var vars = map[string]string{
			Placeholder.CurrentPackage:     currentPkg.PkgPath,
			Placeholder.CurrentPackageName: currentPkg.Name,
			Placeholder.TargetStructName:   cf.TargetStructName,
			Placeholder.SourceStructName:   cf.SourceStructName,
		}

		targetStruct, ok := ctx.Parser().FindStruct(replacePlaceholders(cf.TargetPkgPath, vars), cf.TargetStructName)
		if !ok {
			// log struct not found
			continue
		}

		sourceStruct, ok := ctx.Parser().FindStruct(replacePlaceholders(cf.SourcePkgPath, vars), cf.SourceStructName)
		if !ok {
			// log source struct not found
			continue
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

		if cf.GenerateSourceToTarget {
			toTargetFuncName := replacePlaceholders(cf.SourceToTargetFuncName, vars)

			tv := vars
			tv[Placeholder.FunctionName] = toTargetFuncName
			decorateToTargetFuncName := replacePlaceholders(cf.DecorateFuncName, tv)

			mapFunc := genMapFunc{
				name:              cf.MapperName + "-SourceToTarget",
				funcName:          toTargetFuncName,
				decorateFuncName:  decorateToTargetFuncName,
				targetParamName:   "out",
				targetPkgPath:     cf.TargetPkgPath,
				targetType:        targetStruct.Type,
				targetPointer:     useTargetPointer,
				sourceParamName:   "in",
				sourcePkgPath:     cf.SourcePkgPath,
				sourceType:        sourceStruct.Type,
				sourcePointer:     useSourcePointer,
				targetFieldsIndex: makeFieldsIndex(targetStruct.Fields),
				sourceFieldsIndex: makeFieldsIndex(sourceStruct.Fields),
			}

			fillMapFunc(ctx, &mapFunc, targetStruct.Fields, sourceStruct.Fields, cf.Fields, cf.UseGetter)
			mapFuncs = append(mapFuncs, &mapFunc)
		}

		if cf.GenerateSourceFromTarget {
			fromTargetFuncName := replacePlaceholders(cf.SourceFromTargetFuncName, vars)

			fv := vars
			fv[Placeholder.FunctionName] = fromTargetFuncName
			decorateFromTargetFuncName := replacePlaceholders(cf.DecorateFuncName, fv)

			mapFunc := genMapFunc{
				name:              cf.MapperName + "-TargetToSource",
				funcName:          fromTargetFuncName,
				decorateFuncName:  decorateFromTargetFuncName,
				targetParamName:   "out",
				targetPkgPath:     cf.TargetPkgPath,
				targetType:        sourceStruct.Type,
				targetPointer:     useSourcePointer,
				sourceParamName:   "in",
				sourcePkgPath:     cf.SourcePkgPath,
				sourceType:        targetStruct.Type,
				sourcePointer:     useTargetPointer,
				targetFieldsIndex: makeFieldsIndex(sourceStruct.Fields),
				sourceFieldsIndex: makeFieldsIndex(targetStruct.Fields),
			}

			fillMapFunc(ctx, &mapFunc, sourceStruct.Fields, targetStruct.Fields, cf.Fields.Flip(), cf.UseGetter)
			mapFuncs = append(mapFuncs, &mapFunc)
		}
	}
	return mapFuncs, nil
}

func fillMapFunc(ctx ConverterContext, mapFunc *genMapFunc, targetFields, sourceFields map[string]StructFieldInfo, config FieldConfig, useGetter bool) {
	samePkg := mapFunc.targetPkgPath == mapFunc.sourcePkgPath
	mappedFields := mapFieldNames(targetFields, sourceFields, config, samePkg)
	for target, source := range mappedFields {
		if source == "" {
			if samePkg {
				mapFunc.missingFields = append(mapFunc.missingFields, target)
				continue
			}

			if targetInfo, ok := targetFields[target]; ok && targetInfo.IsExported {
				mapFunc.missingFields = append(mapFunc.missingFields, target)
			}
		}

		ti, ok := targetFields[target]
		if !ok {
			continue
		}
		si, ok := sourceFields[source]
		if !ok {
			continue
		}

		converter, ok := findConverter(ctx, ti.Type, si.Type)
		if !ok {
			mapFunc.unconvertibleFields = append(mapFunc.unconvertibleFields, target)
			continue
		}

		sourceSymbol := newSymbol("in", source, si.Type)
		if useGetter && si.Getter != nil {
			sourceSymbol = sourceSymbol.toGetterSymbol(*si.Getter)
		}

		field := convertibleField{
			index:           ti.Index,
			targetFieldName: target,
			targetSymbol:    newSymbol("out", target, ti.Type),
			sourceFieldName: source,
			sourceSymbol:    sourceSymbol,
			converter:       converter,
		}

		opt := ConverterOption{}

		// this is a run to check that converted code is nil or not if converted code is nil we
		//consider it unconvertible. The main run is in generator... function.
		convertedCode := field.converter.ConvertField(ctx, field.targetSymbol, field.sourceSymbol, opt)
		if convertedCode == nil {
			mapFunc.appendUnconvertibleField(field.targetFieldName)
		}

		mapFunc.mappedFields = append(mapFunc.mappedFields, field)
	}

	slices.SortFunc(mapFunc.mappedFields, func(a, b convertibleField) int {
		return a.index - b.index
	})
}

func mapFieldNames(targetFields, sourceFields map[string]StructFieldInfo, config FieldConfig, samePkg bool) map[string]string {
	result := make(map[string]string)
	for target, targetInfo := range targetFields {
		if !samePkg && !targetInfo.IsExported {
			continue
		}

		result[target] = ""
		if config.ManualMap != nil {
			manualSource, ok := config.ManualMap[target]
			if ok {
				info, have := sourceFields[manualSource]
				if !have {
					continue
				}

				if !samePkg && !info.IsExported {
					continue
				}

				result[target] = manualSource
				continue
			}
		}

		for source, sourceInfo := range sourceFields {
			if !samePkg && !sourceInfo.IsExported {
				continue
			}

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

func hasMissingOrUnconvertibleField(fns []*genMapFunc) bool {
	for _, mf := range fns {
		if len(mf.missingFields) > 0 {
			return true
		}
		if len(mf.unconvertibleFields) > 0 {
			return true
		}
	}
	return false
}

func makeFieldsIndex(in map[string]StructFieldInfo) map[string]int {
	result := make(map[string]int)
	for k, v := range in {
		result[k] = v.Index
	}
	return result
}

func sortFieldsByIndex(input []string, index map[string]int) []string {
	slices.SortFunc(input, func(a, b string) int {
		ai, ok := index[a]
		if !ok {
			ai = math.MaxInt
		}

		bi, ok := index[b]
		if !ok {
			bi = math.MaxInt
		}

		return ai - bi
	})
	return input
}
