package gomappergen

import (
	"go/types"
	"log/slog"

	"github.com/dave/jennifer/jen"
)

type numericConverter struct {
	numericTypes []types.Type
}

func (c *numericConverter) Init(_ Parser, _ Config, _ *slog.Logger) {
	c.numericTypes = []types.Type{
		types.Typ[types.Int],
		types.Typ[types.Int8],
		types.Typ[types.Int16],
		types.Typ[types.Int32],
		types.Typ[types.Int64],
		types.Typ[types.Uint],
		types.Typ[types.Uint8],
		types.Typ[types.Uint16],
		types.Typ[types.Uint32],
		types.Typ[types.Uint64],
		types.Typ[types.Float32],
		types.Typ[types.Float64],
	}
}

func (c *numericConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in numericConverter",
		ShortForm:            "[T number] -> [V number]",
		ShortFormDescription: "Convert between number types (int*, uint*, float*)",
	}
}

func (c *numericConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	if c.isNumeric(sourceType) && c.isNumeric(targetType) {
		return true
	}

	if c.isNumeric(sourceType) {
		_, _, ok := c.lookUpConvertibleFromNumeric(ctx, targetType)
		return ok
	}

	if c.isNumeric(targetType) {
		_, _, ok := c.lookUpConverterToNumeric(ctx, sourceType)
		return ok
	}

	_, _, ok1 := c.lookUpConverterToNumeric(ctx, sourceType)
	_, _, ok2 := c.lookUpConvertibleFromNumeric(ctx, targetType)
	return ok1 && ok2
}

func (c *numericConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		switch {
		case c.isNumeric(target.Type) && c.isNumeric(source.Type):
			code := jen.Add(GeneratorUtil.TypeToJenCode(target.Type)).Params(source.Expr())
			return target.Expr().Op("=").Add(code)

		case c.isNumeric(target.Type):
			oc, numericType, ok := c.lookUpConverterToNumeric(ctx, source.Type)
			if !ok {
				return nil
			}

			// first convert source -> numericType using oc - other converter
			varName := ctx.NextVarName()
			code := jen.Line().Var().Id(varName).Add(GeneratorUtil.TypeToJenCode(numericType)).Line()

			targetSymbol := Symbol{VarName: varName, Type: numericType, Metadata: SymbolMetadata{IsVariable: true, HasZeroValue: true}}
			convertedCode := oc.ConvertField(ctx, targetSymbol, source)
			if convertedCode == nil {
				return nil
			}
			code.Add(convertedCode).Line()

			// then convert numericType -> targetType by casting
			rhs := jen.Add(GeneratorUtil.TypeToJenCode(target.Type)).Params(jen.Id(varName))
			return code.Add(target.Expr().Op("=").Add(rhs))

		case c.isNumeric(source.Type):
			oc, numericType, ok := c.lookUpConvertibleFromNumeric(ctx, target.Type)
			if !ok {
				return nil
			}

			// first convert source.Type to numericType by casting
			varName := ctx.NextVarName()
			rhs := jen.Add(GeneratorUtil.TypeToJenCode(numericType)).Params(source.Expr())
			code := jen.Id(varName).Op(":=").Add(rhs).Line()

			// then convert numericType to target.Type using oc - other converter
			sourceSymbol := Symbol{VarName: varName, Type: numericType}
			convertedCode := oc.ConvertField(ctx, target, sourceSymbol)
			if convertedCode == nil {
				return nil
			}
			return code.Add(convertedCode).Line()

		default:
			bc, beforeNumericType, ok1 := c.lookUpConverterToNumeric(ctx, source.Type)
			ac, afterNumericType, ok2 := c.lookUpConvertibleFromNumeric(ctx, target.Type)
			if !ok1 || !ok2 {
				return nil
			}

			// first convert source.Type -> beforeNumericType using bc - before converter
			beforeVarName := ctx.NextVarName()
			code := jen.Line().Var().Id(beforeVarName).Add(GeneratorUtil.TypeToJenCode(beforeNumericType)).Line()

			targetSymbol := Symbol{VarName: beforeVarName, Type: beforeNumericType, Metadata: SymbolMetadata{IsVariable: true, HasZeroValue: true}}
			convertedCodeBefore := bc.ConvertField(ctx, targetSymbol, source)
			if convertedCodeBefore == nil {
				return nil
			}
			code.Add(convertedCodeBefore).Line()

			// then convert beforeNumericType -> afterNumericType by casting
			afterVarName := ctx.NextVarName()

			rhs := jen.Add(GeneratorUtil.TypeToJenCode(afterNumericType)).Params(jen.Id(beforeVarName))
			code = code.Add(jen.Id(afterVarName).Op(":=").Add(rhs).Line())

			// then convert afterNumericType -> target.Type using ac - after converter
			sourceSymbol := Symbol{VarName: afterVarName, Type: afterNumericType}
			afterConvertedCode := ac.ConvertField(ctx, target, sourceSymbol)
			if afterConvertedCode == nil {
				return nil
			}
			return code.Add(afterConvertedCode).Line()
		}
	})
}

func (c *numericConverter) isNumeric(t types.Type) bool {
	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	const numeric = types.IsInteger | types.IsFloat
	return basic.Info()&numeric != 0
}

func (c *numericConverter) lookUpConverterToNumeric(ctx LookupContext, t types.Type) (Converter, types.Type, bool) {
	for _, n := range c.numericTypes {
		v, _ := ctx.LookUp(c, n, t)
		if v != nil {
			return v, n, true
		}
	}
	return nil, nil, false
}

func (c *numericConverter) lookUpConvertibleFromNumeric(ctx LookupContext, t types.Type) (Converter, types.Type, bool) {
	for _, n := range c.numericTypes {
		v, _ := ctx.LookUp(c, t, n)
		if v != nil {
			return v, n, true
		}
	}
	return nil, nil, false
}

var _ Converter = (*numericConverter)(nil)
