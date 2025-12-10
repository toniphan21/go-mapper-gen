package gomappergen

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

// --- identical type

type identicalTypeConverter struct {
}

func (c *identicalTypeConverter) Print() string {
	return "T -> T  (Direct Value Copy)"
}

func (c *identicalTypeConverter) CanConvert(targetType, sourceType types.Type) bool {
	return types.Identical(targetType, sourceType)
}

func (c *identicalTypeConverter) Before(file *jen.File, target, source Symbol, currentVarCount int) (code jen.Code, nextVarCount int) {
	return nil, currentVarCount
}

func (c *identicalTypeConverter) Assign(file *jen.File, target, source Symbol) jen.Code {
	code := jen.Id(target.VarName).Dot(target.FieldName)
	code = code.Op("=")
	code = code.Id(source.VarName).Dot(source.FieldName)
	return code
}

var _ Converter = (*identicalTypeConverter)(nil)

// --- pointer to type

type typeToPointerConverter struct {
}

func (c *typeToPointerConverter) Print() string {
	return "T -> *T (Address of Source Value). Skips interfaces."
}

func (c *typeToPointerConverter) CanConvert(targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(targetType, sourceType) && !TypeUtil.IsInterface(sourceType)
}

func (c *typeToPointerConverter) Before(file *jen.File, target, source Symbol, currentVarCount int) (code jen.Code, generatedVarCount int) {
	return nil, currentVarCount
}

func (c *typeToPointerConverter) Assign(file *jen.File, target, source Symbol) jen.Code {
	code := jen.Id(target.VarName).Dot(target.FieldName)
	code = code.Op("=")
	code = code.Op("&").Id(source.VarName).Dot(source.FieldName)
	return code
}

var _ Converter = (*typeToPointerConverter)(nil)

// --- type to pointer

type pointerToTypeConverter struct {
}

func (c *pointerToTypeConverter) Print() string {
	return "*T -> T (Nil check and Dereference). Uses zero value if source is nil. Skips interfaces."
}

func (c *pointerToTypeConverter) CanConvert(targetType, sourceType types.Type) bool {
	return TypeUtil.IsPointerOfType(sourceType, targetType) && !TypeUtil.IsInterface(targetType)
}

func (c *pointerToTypeConverter) Before(file *jen.File, target, source Symbol, currentVarCount int) (code jen.Code, generatedVarCount int) {
	return nil, currentVarCount
}

func (c *pointerToTypeConverter) Assign(file *jen.File, target, source Symbol) jen.Code {
	return jen.If(jen.Id(source.VarName).Dot(source.FieldName).Op("==").Nil()).
		BlockFunc(func(g *jen.Group) {
			code := g.Var().Id("v").Add(GeneratorUtil.TypeToJenCode(target.Type)).Line()
			code = code.Id(target.VarName).Dot(target.FieldName).Op("=").Id("v")
		}).
		Else().
		BlockFunc(func(g *jen.Group) {
			code := g.Id(target.VarName).Dot(target.FieldName)
			code = code.Op("=")
			code = code.Op("*").Id(source.VarName).Dot(source.FieldName)
		})
}

var _ Converter = (*pointerToTypeConverter)(nil)
