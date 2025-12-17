package gomappergen

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

type funcConverter struct {
	targetType   types.Type
	sourceType   types.Type
	pkgPath      string
	variableName *string
	funcName     string
}

type funcConverterMatch struct {
	before Converter
	fn     *funcConverter
	after  Converter
}

func (m *funcConverterMatch) CanConvert() bool {
	return m.fn != nil
}

type functionsConverter struct {
	availableFunctions []funcConverter
}

func (c *functionsConverter) Init(parser Parser, config Config) {
	if len(config.ConverterFunctions) == 0 {
		// no-op
		return
	}

	for _, v := range config.ConverterFunctions {
		fn, ok := parser.FindFunction(v.PackagePath, v.TypeName)
		if ok {
			if len(fn.Params) != 1 && len(fn.Results) != 1 {
				continue
			}

			c.availableFunctions = append(c.availableFunctions, funcConverter{
				sourceType: fn.Params[0],
				targetType: fn.Results[0],
				pkgPath:    fn.PackagePath,
				funcName:   fn.Name,
			})
			continue
		}
	}
}

func (c *functionsConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in functionsConverter",
		ShortForm:            "(func(T) -> V)(T)",
		ShortFormDescription: "invoke converter functions",
	}
}

func (c *functionsConverter) CanConvert(ctx ConverterContext, targetType, sourceType types.Type) bool {
	match := c.matchFuncConverter(ctx, targetType, sourceType)

	return match.CanConvert()
}

func (c *functionsConverter) matchFuncConverter(ctx ConverterContext, targetType, sourceType types.Type) funcConverterMatch {
	for _, fn := range c.availableFunctions {
		identicalTarget := types.Identical(fn.targetType, targetType)
		identicalSource := types.Identical(fn.sourceType, sourceType)

		if identicalTarget && identicalSource {
			return funcConverterMatch{fn: &fn}
		}

		before, convertibleSource := ctx.LookUp(c, fn.sourceType, sourceType)
		if identicalTarget && convertibleSource {
			return funcConverterMatch{before: before, fn: &fn, after: nil}
		}

		after, convertibleTarget := ctx.LookUp(c, targetType, fn.targetType)
		if identicalSource && convertibleTarget {
			return funcConverterMatch{before: nil, fn: &fn, after: after}
		}

		if convertibleTarget && convertibleSource {
			return funcConverterMatch{before: before, fn: &fn, after: after}
		}
	}

	return funcConverterMatch{}
}

func (c *functionsConverter) ConvertField(ctx ConverterContext, target, source Symbol, opt ConverterOption) jen.Code {
	return ctx.Run(c, opt, func() jen.Code {
		match := c.matchFuncConverter(ctx, target.Type, source.Type)
		if !match.CanConvert() {
			return nil
		}

		if match.before == nil && match.after == nil {
			return target.Expr().Op("=").Qual(match.fn.pkgPath, match.fn.funcName).Params(source.Expr())
		}

		if match.after == nil {
			varName := ctx.NextVarName()
			code := jen.Var().Id(varName).Add(GeneratorUtil.TypeToJenCode(match.fn.sourceType)).Line()

			// use before convert source.Type -> fn.sourceType
			targetSymbol := Symbol{VarName: varName, Type: match.fn.sourceType}
			ccode := match.before.ConvertField(ctx, targetSymbol, source, opt)
			if ccode == nil {
				return nil
			}
			code = code.Add(ccode).Line()

			// use fn convert fn.sourceType -> target.Type
			code = code.Add(target.Expr().Op("=").Qual(match.fn.pkgPath, match.fn.funcName).Params(jen.Id(varName)))
			return code
		}

		if match.before == nil {
			// use fn convert source -> fn.targetType
			varName := ctx.NextVarName()
			code := jen.Id(varName).Op(":=").Qual(match.fn.pkgPath, match.fn.funcName).Params(source.Expr()).Line()

			// use after convert fn.targetType -> target.Type
			sourceSymbol := Symbol{VarName: varName, Type: match.fn.targetType}
			ccode := match.after.ConvertField(ctx, target, sourceSymbol, opt)
			if ccode == nil {
				return nil
			}
			code.Add(ccode)
			return code
		}

		beforeVarName := ctx.NextVarName()
		code := jen.Var().Id(beforeVarName).Add(GeneratorUtil.TypeToJenCode(match.fn.sourceType)).Line()
		// use before convert source.Type -> fn.sourceType
		targetSymbol := Symbol{VarName: beforeVarName, Type: match.fn.sourceType}
		bCode := match.before.ConvertField(ctx, targetSymbol, source, opt)
		if bCode == nil {
			return nil
		}
		code = code.Add(bCode).Line()

		// use fn convert fn.sourceType -> fn.targetType
		afterVarName := ctx.NextVarName()
		code = code.Add(jen.Id(afterVarName).Op(":=").Qual(match.fn.pkgPath, match.fn.funcName).Params(jen.Id(beforeVarName))).Line()

		// use after convert fn.targetType -> target.Type
		sourceSymbol := Symbol{VarName: afterVarName, Type: match.fn.targetType}
		aCode := match.after.ConvertField(ctx, target, sourceSymbol, opt)
		if aCode == nil {
			return nil
		}
		code.Add(aCode)
		return code
	})
}

var _ Converter = (*functionsConverter)(nil)
