package gomappergen

import (
	"go/types"
	"log/slog"

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

type baseFunctionsConverter struct {
	converter          Converter
	availableFunctions []funcConverter
	strict             bool
}

func (b *baseFunctionsConverter) Init(parser Parser, config Config, _ *slog.Logger) {
	if len(config.ConverterFunctions) == 0 {
		// no-op
		return
	}

	b.availableFunctions = make([]funcConverter, 0)
	for _, v := range config.ConverterFunctions {
		fn, ok := parser.FindFunction(v.PackagePath, v.TypeName)
		if ok {
			if len(fn.Params) != 1 || len(fn.Results) != 1 {
				continue
			}

			b.availableFunctions = append(b.availableFunctions, funcConverter{
				sourceType: fn.Params[0],
				targetType: fn.Results[0],
				pkgPath:    fn.PackagePath,
				funcName:   fn.Name,
			})
			continue
		}

		varFns := parser.FindVariableMethods(v.PackagePath, v.TypeName)
		if len(varFns) > 0 {
			variableName := v.TypeName
			for _, vfn := range varFns {
				if len(vfn.Params) != 1 || len(vfn.Results) != 1 {
					continue
				}

				b.availableFunctions = append(b.availableFunctions, funcConverter{
					sourceType:   vfn.Params[0],
					targetType:   vfn.Results[0],
					variableName: &variableName,
					pkgPath:      vfn.PackagePath,
					funcName:     vfn.Name,
				})
				continue
			}
		}
	}
}

func (b *baseFunctionsConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	match := b.matchFuncConverter(ctx, targetType, sourceType)

	return match.CanConvert()
}

func (b *baseFunctionsConverter) matchFuncConverter(ctx LookupContext, targetType, sourceType types.Type) funcConverterMatch {
	for _, fn := range b.availableFunctions {
		identicalTarget := TypeUtil.IsIdentical(fn.targetType, targetType)
		identicalSource := TypeUtil.IsIdentical(fn.sourceType, sourceType)

		if identicalTarget && identicalSource {
			return funcConverterMatch{fn: &fn}
		}

		if b.strict {
			continue
		}

		before, err := ctx.LookUp(b.converter, fn.sourceType, sourceType)
		convertibleSource := err == nil
		if identicalTarget && convertibleSource {
			return funcConverterMatch{before: before, fn: &fn, after: nil}
		}

		after, err := ctx.LookUp(b.converter, targetType, fn.targetType)
		convertibleTarget := err == nil
		if identicalSource && convertibleTarget {
			return funcConverterMatch{before: nil, fn: &fn, after: after}
		}

		if convertibleTarget && convertibleSource {
			return funcConverterMatch{before: before, fn: &fn, after: after}
		}
	}

	return funcConverterMatch{}
}

func (b *baseFunctionsConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(b.converter, func() jen.Code {
		match := b.matchFuncConverter(ctx, target.Type, source.Type)
		if !match.CanConvert() {
			return nil
		}

		if match.before == nil && match.after == nil {
			code := target.Expr().Op("=")
			if match.fn.variableName != nil {
				return code.Qual(match.fn.pkgPath, *match.fn.variableName).Dot(match.fn.funcName).Params(source.Expr())
			}
			return code.Qual(match.fn.pkgPath, match.fn.funcName).Params(source.Expr())
		}

		if match.after == nil {
			varName := ctx.NextVarName()
			code := jen.Var().Id(varName).Add(GeneratorUtil.TypeToJenCode(match.fn.sourceType)).Line()

			// use before convert source.Type -> fn.sourceType
			targetSymbol := Symbol{VarName: varName, Type: match.fn.sourceType, Metadata: SymbolMetadata{IsVariable: true, HasZeroValue: true}}
			ccode := match.before.ConvertField(ctx, targetSymbol, source)
			if ccode == nil {
				return nil
			}
			code = code.Add(ccode).Line()

			// use fn convert fn.sourceType -> target.Type
			fc := target.Expr().Op("=")
			if match.fn.variableName != nil {
				fc = fc.Qual(match.fn.pkgPath, *match.fn.variableName).Dot(match.fn.funcName).Params(jen.Id(varName))
			} else {
				fc = fc.Qual(match.fn.pkgPath, match.fn.funcName).Params(jen.Id(varName))
			}
			return code.Add(fc)
		}

		if match.before == nil {
			// use fn convert source -> fn.targetType
			varName := ctx.NextVarName()
			code := jen.Id(varName).Op(":=")
			if match.fn.variableName != nil {
				code = code.Qual(match.fn.pkgPath, *match.fn.variableName).Dot(match.fn.funcName).Params(source.Expr()).Line()
			} else {
				code = code.Qual(match.fn.pkgPath, match.fn.funcName).Params(source.Expr()).Line()
			}

			// use after convert fn.targetType -> target.Type
			sourceSymbol := Symbol{VarName: varName, Type: match.fn.targetType}
			ccode := match.after.ConvertField(ctx, target, sourceSymbol)
			if ccode == nil {
				return nil
			}
			code.Add(ccode)
			return code
		}

		beforeVarName := ctx.NextVarName()
		code := jen.Var().Id(beforeVarName).Add(GeneratorUtil.TypeToJenCode(match.fn.sourceType)).Line()
		// use before convert source.Type -> fn.sourceType
		targetSymbol := Symbol{VarName: beforeVarName, Type: match.fn.sourceType, Metadata: SymbolMetadata{IsVariable: true, HasZeroValue: true}}
		bCode := match.before.ConvertField(ctx, targetSymbol, source)
		if bCode == nil {
			return nil
		}
		code = code.Add(bCode).Line()

		// use fn convert fn.sourceType -> fn.targetType
		afterVarName := ctx.NextVarName()
		fc := jen.Id(afterVarName).Op(":=")
		if match.fn.variableName != nil {
			fc = fc.Qual(match.fn.pkgPath, *match.fn.variableName).Dot(match.fn.funcName).Params(jen.Id(beforeVarName))
		} else {
			fc = fc.Qual(match.fn.pkgPath, match.fn.funcName).Params(jen.Id(beforeVarName))
		}
		code = code.Add(fc).Line()

		// use after convert fn.targetType -> target.Type
		sourceSymbol := Symbol{VarName: afterVarName, Type: match.fn.targetType}
		aCode := match.after.ConvertField(ctx, target, sourceSymbol)
		if aCode == nil {
			return nil
		}
		code.Add(aCode)
		return code
	})
}

type functionsConverter struct {
	base baseFunctionsConverter
}

func (c *functionsConverter) Init(parser Parser, config Config, logger *slog.Logger) {
	c.base.strict = false
	c.base.converter = c
	c.base.Init(parser, config, logger)
}

func (c *functionsConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return c.base.CanConvert(ctx, targetType, sourceType)
}

func (c *functionsConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return c.base.ConvertField(ctx, target, source)
}

func (c *functionsConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in functionsConverter",
		ShortForm:            "loosely invoke (T) -> V",
		ShortFormDescription: "invoke converter functions loosely",
	}
}

var _ Converter = (*functionsConverter)(nil)

type strictFunctionsConverter struct {
	base baseFunctionsConverter
}

func (c *strictFunctionsConverter) Init(parser Parser, config Config, logger *slog.Logger) {
	c.base.strict = true
	c.base.converter = c
	c.base.Init(parser, config, logger)
}

func (c *strictFunctionsConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return c.base.CanConvert(ctx, targetType, sourceType)
}

func (c *strictFunctionsConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return c.base.ConvertField(ctx, target, source)
}

func (c *strictFunctionsConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in strictFunctionsConverter",
		ShortForm:            "strictly invoke (T) -> V",
		ShortFormDescription: "invoke converter functions strictly",
	}
}

var _ Converter = (*strictFunctionsConverter)(nil)
