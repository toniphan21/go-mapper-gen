package gomappergen

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
	"golang.org/x/tools/go/packages"
)

type generatorImpl struct {
	parser      Parser
	fileManager FileManager
	logger      *slog.Logger
}

func (g *generatorImpl) Generate(currentPkg *packages.Package, configs []PackageConfig) error {
	for _, cf := range configs {
		file := g.fileManager.MakeJenFile(g.parser, currentPkg, cf)
		if file == nil {
			continue
		}

		if err := generateMapper(g.parser, file, currentPkg, cf, g.logger); err != nil {
			return err
		}
	}
	return nil
}

var _ Generator = (*generatorImpl)(nil)

type convertibleField struct {
	index            int
	targetFieldName  string
	targetSymbol     Symbol
	sourceFieldName  string
	sourceSymbol     Symbol
	converter        Converter
	targetDescriptor Descriptor
	sourceDescriptor Descriptor
	interceptor      FieldInterceptor
}

func (f *convertibleField) PerformConvertField(ctx *converterContext) jen.Code {
	if f.interceptor == nil {
		ctx.resetFieldInterceptor()
		return f.converter.ConvertField(ctx, f.targetSymbol, f.sourceSymbol)
	}
	ctx.setFieldInterceptor(f.interceptor)
	return f.interceptor.InterceptConvertField(f.converter, ctx, f.targetSymbol, f.sourceSymbol)
}

type genMapFunc struct {
	name                string
	funcName            string
	decorateFuncName    string
	targetPkgPath       string
	targetParamName     string
	targetStruct        *StructInfo
	targetPointer       bool
	sourcePkgPath       string
	sourceParamName     string
	sourceStruct        *StructInfo
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
		params = append(params, jen.Id(mf.sourceParamName).Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceStruct.Type))))
	} else {
		params = append(params, jen.Id(mf.sourceParamName).Add(GeneratorUtil.TypeToJenCode(mf.sourceStruct.Type)))
	}

	if mf.targetPointer {
		result = append(result, jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type)))
	} else {
		result = append(result, jen.Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type)))
	}

	return params, result
}

func (mf *genMapFunc) appendUnconvertibleField(field string) {
	mf.unconvertibleFields = append(mf.unconvertibleFields, field)
}

func generateMapper(parser Parser, file *jen.File, currentPkg *packages.Package, config PackageConfig, logger *slog.Logger) error {
	ctx := &converterContext{
		Context:       context.Background(),
		lookupContext: emptyLookupContext(logger),
		jenFile:       file,
		parser:        parser,
	}

	mapFuncs, err := collectMapFuncs(ctx, currentPkg, config, logger)
	if err != nil {
		return err
	}

	if len(mapFuncs) == 0 {
		logger.Info("\tthere is no structs matched with configuration.")
		return nil
	}

	slices.SortFunc(mapFuncs, func(a, b *genMapFunc) int {
		return strings.Compare(a.name, b.name)
	})

	logger.Info(fmt.Sprintf("\tthere are %d map functions matched with configuration.", len(mapFuncs)))
	for _, mf := range mapFuncs {
		logger.Info(fmt.Sprintf("\t\t- %s(%s) %s", util.ColorBlue(mf.funcName), mf.sourceStruct.Type.String(), mf.targetStruct.Type.String()))
	}

	switch config.Mode {
	case ModeFunctions:
		logger.Info("\tgenerating mode functions...")
		generateMapperFunctions(ctx, currentPkg, config, mapFuncs)

	default:
		logger.Info("\tgenerating mode types...")
		generateMapperInterface(file, currentPkg, config, mapFuncs, logger)
		generateDecoratorInterface(ctx, config, mapFuncs, logger)
		generateMapperConstructor(ctx, config, mapFuncs, logger)
		generateMapperImplementation(ctx, config, mapFuncs, logger)
		generateDecoratorNoOp(ctx, config, mapFuncs, logger)
		generateCompileTimeCheck(file, config, mapFuncs, logger)
	}
	logger.Info("\tfinished")

	return nil
}

func generateMapperFunctions(ctx *converterContext, currentPkg *packages.Package, config PackageConfig, mapFuncs []*genMapFunc) {
	file := ctx.JenFile()

	for _, mf := range mapFuncs {
		ctx.resetVarCount()

		params, results := mf.paramsAndResults()

		useDecorator := len(mf.missingFields) > 0 || len(mf.unconvertibleFields) > 0
		if config.DecoratorMode == DecoratorModeAlways {
			useDecorator = true
		} else if config.DecoratorMode == DecoratorModeNever {
			useDecorator = false
		}

		if useDecorator {
			params = append(params,
				jen.Id("decorators").Op("...").Func().
					Params(
						jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceStruct.Type)),
						jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type)),
					),
			)
		}

		var body []jen.Code
		body = append(body, jen.Var().Id(mf.targetParamName).Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type)).Line())

		for _, field := range mf.mappedFields {
			ctx.resetLookupContext(field.targetDescriptor, field.sourceDescriptor)
			convertedCode := field.PerformConvertField(ctx)
			if convertedCode != nil {
				body = append(body, convertedCode)
			}
		}

		mau := makeMissingAndUnconvertibleFields(mf)
		if len(mau) > 0 {
			body = append(body, jen.Line())
			body = append(body, mau...)
		}

		if useDecorator {
			var decoratorParams []jen.Code

			if mf.sourcePointer {
				decoratorParams = append(decoratorParams, jen.Id(mf.sourceParamName))
			} else {
				decoratorParams = append(decoratorParams, jen.Op("&").Id(mf.sourceParamName))
			}
			decoratorParams = append(decoratorParams, jen.Op("&").Id(mf.targetParamName))

			code := jen.
				For(jen.List(jen.Id("_"), jen.Id("decorate")).Op(":=").Range().Id("decorators")).
				Block(jen.Id("decorate").Call(decoratorParams...))

			body = append(body, jen.Line())
			body = append(body, code)
		}

		if mf.targetPointer {
			body = append(body, jen.Line().Return(jen.Op("&").Id(mf.targetParamName)))
		} else {
			body = append(body, jen.Line().Return(jen.Id(mf.targetParamName)))
		}

		if config.GenerateGoDoc {
			comment := fmt.Sprintf(
				"%v converts a %v value into a %v value.",
				mf.funcName,
				GeneratorUtil.SimpleNameWithPkg(currentPkg, mf.sourceStruct.Type),
				GeneratorUtil.SimpleNameWithPkg(currentPkg, mf.targetStruct.Type),
			)
			file.Comment(comment)
		}

		file.Func().
			Id(mf.funcName).
			Params(params...).
			Params(results...).
			Block(body...).
			Line()
	}
}

func generateMapperInterface(file *jen.File, currentPkg *packages.Package, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	var signatures []jen.Code

	for _, mf := range mapFuncs {
		params, results := mf.paramsAndResults()

		if config.GenerateGoDoc {
			comment := fmt.Sprintf(
				"%v converts a %v value into a %v value.",
				mf.funcName,
				GeneratorUtil.SimpleNameWithPkg(currentPkg, mf.sourceStruct.Type),
				GeneratorUtil.SimpleNameWithPkg(currentPkg, mf.targetStruct.Type),
			)

			signatures = append(signatures, GeneratorUtil.WrapComment(comment))
		}
		signatures = append(signatures, jen.Id(mf.funcName).Params(params...).Params(results...).Line())
	}

	file.Type().Id(config.InterfaceName).Interface(signatures...).Line().Line()
	logger.Info(fmt.Sprintf("\tgenerated interface %s", util.ColorBlue(config.InterfaceName)))
}

func generateDecoratorInterface(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	if !shouldUseDecorator(mapFuncs, config) {
		return
	}

	var signatures []jen.Code

	for _, mf := range mapFuncs {
		var params []jen.Code
		params = append(params, jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceStruct.Type))))
		params = append(params, jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type))))

		signatures = append(signatures, jen.Id(mf.decorateFuncName).Params(params...).Params().Line())
	}

	ctx.jenFile.Type().Id(config.DecoratorInterfaceName).Interface(signatures...).Line().Line()
	logger.Info(fmt.Sprintf("\tgenerated decorator interface %s", util.ColorBlue(config.DecoratorInterfaceName)))
}

func generateMapperConstructor(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	var params, body []jen.Code

	if shouldUseDecorator(mapFuncs, config) {
		params = append(params, jen.Id("decorator").Id(config.DecoratorInterfaceName))
		body = append(body, jen.Return(jen.Op("&").Id(config.ImplementationName).Values(jen.DictFunc(func(d jen.Dict) {
			d[jen.Id("decorator")] = jen.Id("decorator")
		}))))
	} else {
		body = append(body, jen.Return(jen.Op("&").Id(config.ImplementationName).Block(nil)))
	}

	ctx.jenFile.Func().Id(config.ConstructorName).Params(params...).Params(jen.Id(config.InterfaceName)).Block(body...)
	logger.Info(fmt.Sprintf("\tgenerated constructor %s", util.ColorBlue(config.ConstructorName)))
}

func generateMapperImplementation(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	file := ctx.JenFile()
	if shouldUseDecorator(mapFuncs, config) {
		file.Type().Id(config.ImplementationName).Struct(
			jen.Id("decorator").Add(jen.Id(config.DecoratorInterfaceName)),
		).Line()
	} else {
		file.Type().Id(config.ImplementationName).Struct().Line()
	}

	for _, mf := range mapFuncs {
		ctx.resetVarCount()

		params, results := mf.paramsAndResults()

		var body []jen.Code
		body = append(body, jen.Var().Id(mf.targetParamName).Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type)).Line())

		for _, field := range mf.mappedFields {
			ctx.resetLookupContext(field.targetDescriptor, field.sourceDescriptor)
			convertedCode := field.PerformConvertField(ctx)
			if convertedCode != nil {
				body = append(body, convertedCode)
			}
		}

		shouldEmitDecoratorComment := len(mf.missingFields) > 0 || len(mf.unconvertibleFields) > 0
		shouldEmitDecoratorCall := shouldEmitDecoratorComment
		if config.DecoratorMode == DecoratorModeAlways {
			shouldEmitDecoratorCall = true
		} else if config.DecoratorMode == DecoratorModeNever {
			shouldEmitDecoratorCall = false
		}

		if shouldEmitDecoratorCall {
			var decoratorParams []jen.Code

			if mf.sourcePointer {
				decoratorParams = append(decoratorParams, jen.Id(mf.sourceParamName))
			} else {
				decoratorParams = append(decoratorParams, jen.Op("&").Id(mf.sourceParamName))
			}
			decoratorParams = append(decoratorParams, jen.Op("&").Id(mf.targetParamName))

			body = append(body, jen.Line())
			body = append(body, jen.If(jen.Id("m").Dot("decorator").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
				g.Id("m").Dot("decorator").Dot(mf.decorateFuncName).Params(decoratorParams...)
			}))

			// the decorator call is needed and has missing comments already
			shouldEmitDecoratorComment = false
		}

		if shouldEmitDecoratorComment {
			mau := makeMissingAndUnconvertibleFields(mf)
			if len(mau) > 0 {
				body = append(body, jen.Line())
				body = append(body, mau...)
			}
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
	}
	logger.Info(fmt.Sprintf("\tgenerated implementation %s", util.ColorBlue(config.ImplementationName)))
}

func generateDecoratorNoOp(ctx *converterContext, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	if !shouldUseDecorator(mapFuncs, config) || config.DecoratorNoOpName == "" {
		return
	}

	ctx.jenFile.Type().Id(config.DecoratorNoOpName).Struct().Line()

	for _, mf := range mapFuncs {
		ctx.resetVarCount()

		var body []jen.Code
		var params []jen.Code
		params = append(params, jen.Id("in").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.sourceStruct.Type))))
		params = append(params, jen.Id("out").Add(jen.Op("*").Add(GeneratorUtil.TypeToJenCode(mf.targetStruct.Type))))

		mau := makeMissingAndUnconvertibleFields(mf)
		if len(mau) > 0 {
			body = append(body, mau...)
		}

		ctx.JenFile().Func().
			Params(jen.Id("d").Op("*").Id(config.DecoratorNoOpName)).
			Id(mf.decorateFuncName).
			Params(params...).
			Block(body...).
			Line()
	}
	logger.Info(fmt.Sprintf("\tgenerated noop decorator %s", util.ColorBlue(config.DecoratorNoOpName)))
}

func generateCompileTimeCheck(file *jen.File, config PackageConfig, mapFuncs []*genMapFunc, logger *slog.Logger) {
	file.Var().Id("_").Id(config.InterfaceName).Op("=").Parens(jen.Op("*").Id(config.ImplementationName)).Parens(jen.Nil())

	if !shouldUseDecorator(mapFuncs, config) || config.DecoratorNoOpName == "" {
		logger.Info(fmt.Sprintf("\tgenerated compile time check"))
		return
	}
	file.Var().Id("_").Id(config.DecoratorInterfaceName).Op("=").Parens(jen.Op("*").Id(config.DecoratorNoOpName)).Parens(jen.Nil())
	logger.Info(fmt.Sprintf("\tgenerated compile time check"))
}

func makeMissingAndUnconvertibleFields(mf *genMapFunc) []jen.Code {
	var code []jen.Code
	hasMissingField := len(mf.missingFields) > 0
	hasUnconvertibleField := len(mf.unconvertibleFields) > 0

	if hasMissingField {
		code = append(code, jen.Comment("Fields that could not be mapped:"))

		fields := sortFieldsByIndex(mf.missingFields, mf.targetFieldsIndex)
		for _, field := range fields {
			code = append(code, jen.Comment("out."+field+" = "))
		}
	}

	if hasMissingField && hasUnconvertibleField {
		code = append(code, jen.Line())
	}

	if hasUnconvertibleField {
		code = append(code, jen.Comment("Fields that could not be converted (no suitable converter found):"))

		fields := sortFieldsByIndex(mf.unconvertibleFields, mf.targetFieldsIndex)
		for _, field := range fields {
			code = append(code, jen.Comment("out."+field+" = "))
		}
	}
	return code
}

func collectMapFuncs(ctx *converterContext, currentPkg *packages.Package, config PackageConfig, logger *slog.Logger) ([]*genMapFunc, error) {
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
			logger.Warn(
				"\tcannot find target struct",
				slog.String("target_struct_name", cf.TargetStructName),
				slog.String("target_pkg", cf.TargetPkgPath),
			)
			continue
		}

		sourceStruct, ok := ctx.Parser().FindStruct(replacePlaceholders(cf.SourcePkgPath, vars), cf.SourceStructName)
		if !ok {
			logger.Warn(
				"\tcould not find source struct",
				slog.String("source_struct_name", cf.SourceStructName),
				slog.String("source_pkg", cf.SourcePkgPath),
			)
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
				targetStruct:      &targetStruct,
				targetPointer:     useTargetPointer,
				sourceParamName:   "in",
				sourcePkgPath:     cf.SourcePkgPath,
				sourceStruct:      &sourceStruct,
				sourcePointer:     useSourcePointer,
				targetFieldsIndex: makeFieldsIndex(targetStruct.Fields),
				sourceFieldsIndex: makeFieldsIndex(sourceStruct.Fields),
			}

			fillMapFunc(ctx, &mapFunc, targetStruct.Fields, sourceStruct.Fields, cf.Fields, cf.UseGetter, cf.TargetFieldInterceptors)
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
				targetStruct:      &sourceStruct,
				targetPointer:     useSourcePointer,
				sourceParamName:   "in",
				sourcePkgPath:     cf.SourcePkgPath,
				sourceStruct:      &targetStruct,
				sourcePointer:     useTargetPointer,
				targetFieldsIndex: makeFieldsIndex(sourceStruct.Fields),
				sourceFieldsIndex: makeFieldsIndex(targetStruct.Fields),
			}

			fillMapFunc(ctx, &mapFunc, sourceStruct.Fields, targetStruct.Fields, cf.Fields.Flip(), cf.UseGetter, cf.SourceFieldInterceptors)
			mapFuncs = append(mapFuncs, &mapFunc)
		}
	}
	return mapFuncs, nil
}

func fillMapFunc(
	ctx *converterContext,
	mapFunc *genMapFunc,
	targetFields, sourceFields map[string]StructFieldInfo,
	config FieldConfig,
	useGetter bool,
	interceptors map[string]FieldInterceptor,
) {
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
		targetDescriptor := Descriptor{structInfo: mapFunc.targetStruct, structFieldInfo: &ti}
		sourceDescriptor := Descriptor{structInfo: mapFunc.sourceStruct, structFieldInfo: &si}

		converter, ok := findConverter(targetDescriptor, sourceDescriptor, ctx.Logger())
		if !ok {
			mapFunc.unconvertibleFields = append(mapFunc.unconvertibleFields, target)
			continue
		}

		sourceSymbol := newSymbol("in", source, si.Type)
		if useGetter && si.Getter != nil {
			sourceSymbol = sourceSymbol.toGetterSymbol(*si.Getter)
		}

		var interceptor FieldInterceptor
		if interceptors != nil {
			interceptor = interceptors[target]
			if interceptor != nil {
				interceptor.Init(ctx.parser, ctx.Logger())
			}
		}

		field := convertibleField{
			index:            ti.Index,
			targetFieldName:  target,
			targetSymbol:     newSymbolWithMetadata("out", target, ti.Type, SymbolMetadata{HasZeroValue: true}),
			sourceFieldName:  source,
			sourceSymbol:     sourceSymbol,
			converter:        converter,
			targetDescriptor: targetDescriptor,
			sourceDescriptor: sourceDescriptor,
			interceptor:      interceptor,
		}

		// this is a run to check that converted code is nil or not if converted code is nil we
		// consider it unconvertible. The main run is in generator... function.
		ctx.resetLookupContext(targetDescriptor, sourceDescriptor)
		convertedCode := field.PerformConvertField(ctx)
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

func shouldUseDecorator(fns []*genMapFunc, cf PackageConfig) bool {
	if cf.DecoratorMode == DecoratorModeAlways {
		return true
	}

	if cf.DecoratorMode == DecoratorModeNever {
		return false
	}

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
