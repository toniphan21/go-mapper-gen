package gomappergen

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

// --- identical type

type identicalTypeConverter struct {
}

func (c *identicalTypeConverter) Init(_ Parser, _ Config) {
	// no-op
}

func (c *identicalTypeConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in identicalTypeConverter",
		ShortForm:            "T -> T",
		ShortFormDescription: "direct value copy",
	}
}

func (c *identicalTypeConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return TypeUtil.IsIdentical(targetType, sourceType)
}

func (c *identicalTypeConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		return target.Expr().Op("=").Add(source.Expr())
	})
}

var _ Converter = (*identicalTypeConverter)(nil)

// --- pointer to type

type typeToPointerConverter struct {
}

func (c *typeToPointerConverter) Init(_ Parser, _ Config) {
	// no-op
}

func (c *typeToPointerConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in typeToPointerConverter",
		ShortForm:            "T -> *T",
		ShortFormDescription: "address-of; skipped for interface types",
	}
}

func (c *typeToPointerConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(targetType, sourceType) && !TypeUtil.IsInterface(sourceType)
}

func (c *typeToPointerConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		return target.Expr().Op("=").Op("&").Add(source.Expr())
	})
}

var _ Converter = (*typeToPointerConverter)(nil)

// --- type to pointer

type pointerToTypeConverter struct {
}

func (c *pointerToTypeConverter) Init(_ Parser, _ Config) {
	// no-op
}

func (c *pointerToTypeConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in pointerToTypeConverter",
		ShortForm:            "*T -> T",
		ShortFormDescription: "nil-check + dereference; uses zero value when nil; skipped for interface types",
	}
}

func (c *pointerToTypeConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(sourceType, targetType) && !TypeUtil.IsInterface(targetType)
}

func (c *pointerToTypeConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		code := jen.If(source.Expr().Op("!=").Nil()).
			BlockFunc(func(g *jen.Group) {
				gc := g.Add(target.Expr())
				gc = gc.Op("=")
				gc = gc.Op("*").Add(source.Expr())
			})

		if !target.Metadata.HasZeroValue {
			code = code.Else().BlockFunc(func(g *jen.Group) {
				gc := g.Var().Id("zero").Add(GeneratorUtil.TypeToJenCode(target.Type)).Line()
				gc = gc.Add(target.Expr()).Op("=").Id("zero")
			})
		}
		return code
	})
}

var _ Converter = (*pointerToTypeConverter)(nil)

// --- slice

type sliceConverter struct {
}

func (c *sliceConverter) Init(_ Parser, _ Config) {
	// no-op
}

func (c *sliceConverter) Info() ConverterInfo {
	return ConverterInfo{
		Name:                 "built-in sliceConverter",
		ShortForm:            "[]T -> []V",
		ShortFormDescription: "slice conversion; requires converter for T -> V",
	}
}

func (c *sliceConverter) findTypeConverter(ctx LookupContext, targetType, sourceType types.Type) (Converter, types.Type, types.Type, bool) {
	ts, ok := TypeUtil.IsSlice(targetType)
	if !ok {
		return nil, nil, nil, false
	}

	ss, ok := TypeUtil.IsSlice(sourceType)
	if !ok {
		return nil, nil, nil, false
	}

	other, _ := ctx.LookUp(c, ts, ss)
	if other == nil {
		return nil, nil, nil, false
	}
	return other, ts, ss, true
}

func (c *sliceConverter) CanConvert(ctx LookupContext, targetType, sourceType types.Type) bool {
	_, _, _, ok := c.findTypeConverter(ctx, targetType, sourceType)
	return ok
}

func (c *sliceConverter) ConvertField(ctx ConverterContext, target, source Symbol) jen.Code {
	return ctx.Run(c, func() jen.Code {
		other, ts, ss, ok := c.findTypeConverter(ctx, target.Type, source.Type)
		if !ok {
			return nil
		}

		targetSymbol := target.ToIndexedSymbol("i")
		targetSymbol.Type = ts
		sourceSymbol := Symbol{VarName: "v", Type: ss}
		convertCode := other.ConvertField(ctx, targetSymbol, sourceSymbol)
		if convertCode == nil {
			return nil
		}

		code := jen.If(source.Expr().Op("==").Nil()).BlockFunc(func(g *jen.Group) {
			g.Add(target.Expr()).Op("=").Nil()
		})
		code = code.Else().BlockFunc(func(g *jen.Group) {
			gc := g.Add(target.Expr()).Op("=").Make(
				GeneratorUtil.TypeToJenCode(target.Type),
				jen.Len(source.Expr()),
			).Line()

			gc = gc.For(jen.List(jen.Id("i"), jen.Id("v")).Op(":=").Range().Add(source.Expr())).Block(
				convertCode,
			)
		})

		return code
	})
}

var _ Converter = (*sliceConverter)(nil)
