package gomappergen

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

const typeShortHandBuffer = 15

// --- identical type

type identicalTypeConverter struct {
}

func (c *identicalTypeConverter) Print() string {
	return fmt.Sprintf("%-*s direct value copy", typeShortHandBuffer, "T -> T")
}

func (c *identicalTypeConverter) CanConvert(targetType, sourceType types.Type) bool {
	return types.Identical(targetType, sourceType)
}

func (c *identicalTypeConverter) Before(file *jen.File, target, source Symbol, currentVarCount int, opt ConverterOption) (code jen.Code, nextVarCount int) {
	return nil, currentVarCount
}

func (c *identicalTypeConverter) Assign(file *jen.File, target, source Symbol, opt ConverterOption) jen.Code {
	return target.Expr().Op("=").Add(source.Expr())
}

var _ Converter = (*identicalTypeConverter)(nil)

// --- pointer to type

type typeToPointerConverter struct {
}

func (c *typeToPointerConverter) Print() string {
	return fmt.Sprintf("%-*s address-of; skipped for interface types", typeShortHandBuffer, "T -> *T")
}

func (c *typeToPointerConverter) CanConvert(targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(targetType, sourceType) && !TypeUtil.IsInterface(sourceType)
}

func (c *typeToPointerConverter) Before(file *jen.File, target, source Symbol, currentVarCount int, opt ConverterOption) (code jen.Code, generatedVarCount int) {
	return nil, currentVarCount
}

func (c *typeToPointerConverter) Assign(file *jen.File, target, source Symbol, opt ConverterOption) jen.Code {
	code := target.Expr().Op("=").Op("&").Add(source.Expr())
	return GeneratorUtil.GenerateWithConverterOption(code, opt, "built-in typeToPointerConverter.Assign()")
}

var _ Converter = (*typeToPointerConverter)(nil)

// --- type to pointer

type pointerToTypeConverter struct {
}

func (c *pointerToTypeConverter) Print() string {
	return fmt.Sprintf("%-*s nil-check + dereference; uses zero value when nil; skipped for interface types", typeShortHandBuffer, "*T -> T")
}

func (c *pointerToTypeConverter) CanConvert(targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(sourceType, targetType) && !TypeUtil.IsInterface(targetType)
}

func (c *pointerToTypeConverter) Before(file *jen.File, target, source Symbol, currentVarCount int, opt ConverterOption) (code jen.Code, generatedVarCount int) {
	return nil, currentVarCount
}

func (c *pointerToTypeConverter) Assign(file *jen.File, target, source Symbol, opt ConverterOption) jen.Code {
	code := jen.If(source.Expr().Op("==").Nil()).
		BlockFunc(func(g *jen.Group) {
			code := g.Var().Id("zero").Add(GeneratorUtil.TypeToJenCode(target.Type)).Line()
			code = code.Add(target.Expr()).Op("=").Id("zero")
		}).
		Else().
		BlockFunc(func(g *jen.Group) {
			code := g.Add(target.Expr())
			code = code.Op("=")
			code = code.Op("*").Add(source.Expr())
		})
	return GeneratorUtil.GenerateWithConverterOption(code, opt, "built-in pointerToTypeConverter.Assign()")
}

var _ Converter = (*pointerToTypeConverter)(nil)

// --- slice

type sliceConverter struct {
}

func (s *sliceConverter) Print() string {
	return fmt.Sprintf("%-*s slice conversion; requires converter for T -> V", typeShortHandBuffer, "[]T -> []V")
}

func (s *sliceConverter) findTypeConverter(targetType, sourceType types.Type) (Converter, types.Type, types.Type, bool) {
	ts, ok := TypeUtil.IsSlice(targetType)
	if !ok {
		return nil, nil, nil, false
	}

	ss, ok := TypeUtil.IsSlice(sourceType)
	if !ok {
		return nil, nil, nil, false
	}

	c, have := LookUpConverter(s, ts, ss)
	if !have {
		return nil, nil, nil, false
	}
	return c, ts, ss, true
}

func (s *sliceConverter) CanConvert(targetType, sourceType types.Type) bool {
	_, _, _, ok := s.findTypeConverter(targetType, sourceType)
	return ok
}

func (s *sliceConverter) Before(file *jen.File, target, source Symbol, currentVarCount int, opt ConverterOption) (code jen.Code, nextVarCount int) {
	return nil, currentVarCount
}

func (s *sliceConverter) Assign(file *jen.File, target, source Symbol, opt ConverterOption) jen.Code {
	c, ts, ss, ok := s.findTypeConverter(target.Type, source.Type)
	if !ok {
		return nil
	}

	targetSymbol := target.ToIndexedSymbol("i")
	targetSymbol.Type = ts
	sourceSymbol := Symbol{VarName: "v", Type: ss}
	convertCode := c.Assign(file, targetSymbol, sourceSymbol, opt)
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
	return GeneratorUtil.GenerateWithConverterOption(code, opt, "built-in sliceConverter.Assign()")
}

var _ Converter = (*sliceConverter)(nil)
