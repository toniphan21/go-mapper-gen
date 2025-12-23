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

	c.availableFunctions = make([]funcConverter, 0)
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

		varFns := parser.FindVariableMethods(v.PackagePath, v.TypeName)
		if len(varFns) > 0 {
			variableName := v.TypeName
			for _, vfn := range varFns {
				if len(vfn.Params) != 1 && len(vfn.Results) != 1 {
					continue
				}

				c.availableFunctions = append(c.availableFunctions, funcConverter{
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

func (c *functionsConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in functionsConverter",
		ShortForm:            "(func(T) -> V)(T)",
		ShortFormDescription: "invoke converter functions",
	}
}

func (c *functionsConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	match := c.matchFuncConverter(ctx, targetType, sourceType)

	return match.CanConvert()
}

func (c *functionsConverter) matchFuncConverter(ctx LookupContext, targetType, sourceType types.Type) funcConverterMatch {
	for _, fn := range c.availableFunctions {
		identicalTarget := TypeUtil.IsIdentical(fn.targetType, targetType)
		identicalSource := TypeUtil.IsIdentical(fn.sourceType, sourceType)

		if identicalTarget && identicalSource {
			return funcConverterMatch{fn: &fn}
		}

		before, err := ctx.LookUp(c, fn.sourceType, sourceType)
		convertibleSource := err == nil
		if identicalTarget && convertibleSource {
			return funcConverterMatch{before: before, fn: &fn, after: nil}
		}

		after, err := ctx.LookUp(c, targetType, fn.targetType)
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

func (c *functionsConverter) ConvertField(ctx ConverterContext, target, source Symbol, opt ConverterOption) jen.Code {
	return ctx.Run(c, opt, func() jen.Code {
		match := c.matchFuncConverter(ctx, target.Type, source.Type)
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
			targetSymbol := Symbol{VarName: varName, Type: match.fn.sourceType}
			ccode := match.before.ConvertField(ctx, targetSymbol, source, opt)
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
		fc := jen.Id(afterVarName).Op(":=")
		if match.fn.variableName != nil {
			fc = fc.Qual(match.fn.pkgPath, *match.fn.variableName).Dot(match.fn.funcName).Params(jen.Id(beforeVarName))
		} else {
			fc = fc.Qual(match.fn.pkgPath, match.fn.funcName).Params(jen.Id(beforeVarName))
		}
		code = code.Add(fc).Line()

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
