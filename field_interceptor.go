package gomappergen

import (
	"github.com/dave/jennifer/jen"
)

const nilIfZeroType = "nil-if-zero"

type FieldInterceptorProvider interface {
	MakeFieldInterceptor(typ string, options map[string]any) FieldInterceptor
}

type defaultFieldInterceptorProvider struct{}

func (d *defaultFieldInterceptorProvider) MakeFieldInterceptor(typ string, options map[string]any) FieldInterceptor {
	switch typ {
	case nilIfZeroType:
		return BuiltinFieldInterceptor.NilIfZero
	default:
		return nil
	}
}

var _ FieldInterceptorProvider = (*defaultFieldInterceptorProvider)(nil)

// ---

type builtinFieldInterceptor struct {
	NilIfZero FieldInterceptor
}

var BuiltinFieldInterceptor = builtinFieldInterceptor{
	NilIfZero: &nilIfZeroFieldInterceptor{},
}

func DefaultFieldInterceptorProvider() FieldInterceptorProvider {
	return &defaultFieldInterceptorProvider{}
}

// ---

type FieldInterceptor interface {
	GetType() string

	GetOptions() map[string]any

	InterceptConvertField(converter Converter, ctx ConverterContext, target, source Symbol) jen.Code
}

type nilIfZeroFieldInterceptor struct {
}

func (n *nilIfZeroFieldInterceptor) InterceptConvertField(converter Converter, ctx ConverterContext, target, source Symbol) jen.Code {
	c, ok := converter.(*typeToPointerConverter)
	if !ok {
		return converter.ConvertField(ctx, target, source)
	}

	convertedCode := c.ConvertField(ctx, target, source)
	if convertedCode == nil {
		return nil
	}

	varName := ctx.NextVarName()
	code := jen.Var().Id(varName).Add(GeneratorUtil.TypeToJenCode(source.Type)).Line()
	code = code.If(source.Expr().Op("==").Id(varName)).
		BlockFunc(func(g *jen.Group) {
			g.Add(target.Expr()).Op("=").Nil()
		})

	code = code.Else().BlockFunc(func(g *jen.Group) {
		g.Add(convertedCode)
	})
	return code
}

func (n *nilIfZeroFieldInterceptor) GetType() string {
	return nilIfZeroType
}

func (n *nilIfZeroFieldInterceptor) GetOptions() map[string]any {
	return nil
}

var _ FieldInterceptor = (*nilIfZeroFieldInterceptor)(nil)
