package gomappergen

import (
	"fmt"
	"go/types"
	"log/slog"

	"github.com/dave/jennifer/jen"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

const nilIfZeroType = "nil-if-zero"
const useFunctionType = "use-function"

type FieldInterceptor interface {
	GetType() string

	GetOptions() map[string]any

	Init(parser Parser, logger *slog.Logger)

	InterceptCanConvert(converter Converter, ctx LookupContext, targetType, sourceType types.Type) bool

	InterceptConvertField(converter Converter, ctx ConverterContext, target, source Symbol) jen.Code
}

type FieldInterceptorProvider interface {
	MakeFieldInterceptor(typ string, options map[string]any) FieldInterceptor
}

type defaultFieldInterceptorProvider struct{}

func (d *defaultFieldInterceptorProvider) MakeFieldInterceptor(typ string, options map[string]any) FieldInterceptor {
	switch typ {
	case nilIfZeroType:
		return BuiltinFieldInterceptor.NilIfZero
	case useFunctionType:
		s, ok := options["symbol"]
		if !ok {
			return nil
		}
		symbol, ok := s.(string)
		if !ok {
			return nil
		}

		var method string
		m, ok := options["method"]
		if ok {
			method, ok = m.(string)
			if !ok {
				return nil
			}
		}

		if method == "" {
			return BuiltinFieldInterceptor.UseFunction(symbol)
		}
		return BuiltinFieldInterceptor.UseMethod(symbol, method)

	default:
		return nil
	}
}

var _ FieldInterceptorProvider = (*defaultFieldInterceptorProvider)(nil)

// ---

type builtinFieldInterceptor struct {
	NilIfZero   FieldInterceptor
	UseFunction func(symbol string) FieldInterceptor
	UseMethod   func(variableSymbol string, methodName string) FieldInterceptor
}

var BuiltinFieldInterceptor = builtinFieldInterceptor{
	NilIfZero: &nilIfZeroFieldInterceptor{},
	UseFunction: func(symbol string) FieldInterceptor {
		return &useFunctionFieldInterceptor{
			symbol: symbol,
		}
	},
	UseMethod: func(variableSymbol string, methodName string) FieldInterceptor {
		return &useFunctionFieldInterceptor{
			symbol: variableSymbol,
			method: methodName,
		}
	},
}

func DefaultFieldInterceptorProvider() FieldInterceptorProvider {
	return &defaultFieldInterceptorProvider{}
}

// ---

type nilIfZeroFieldInterceptor struct {
}

func (i *nilIfZeroFieldInterceptor) Init(_ Parser, _ *slog.Logger) {
	// no-op
}

func (i *nilIfZeroFieldInterceptor) InterceptCanConvert(converter Converter, ctx LookupContext, targetType, sourceType types.Type) bool {
	return converter.CanConvert(ctx, targetType, sourceType)
}

func (i *nilIfZeroFieldInterceptor) InterceptConvertField(converter Converter, ctx ConverterContext, target, source Symbol) jen.Code {
	c, ok := converter.(*typeToPointerConverter)
	if !ok {
		return converter.ConvertField(ctx, target, source)
	}

	convertedCode := c.ConvertField(ctx, target, source)
	if convertedCode == nil {
		return nil
	}

	if target.Metadata.HasZeroValue {
		varName := ctx.NextVarName()
		code := jen.Var().Id(varName).Add(GeneratorUtil.TypeToJenCode(source.Type)).Line()
		code = code.If(source.Expr().Op("!=").Id(varName)).
			BlockFunc(func(g *jen.Group) {
				g.Add(convertedCode)
			})
		return code
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

func (i *nilIfZeroFieldInterceptor) GetType() string {
	return nilIfZeroType
}

func (i *nilIfZeroFieldInterceptor) GetOptions() map[string]any {
	return nil
}

var _ FieldInterceptor = (*nilIfZeroFieldInterceptor)(nil)

// ---

type useFunctionFieldInterceptor struct {
	symbol   string
	method   string
	function *funcConverter
}

func (i *useFunctionFieldInterceptor) GetType() string {
	return useFunctionType
}

func (i *useFunctionFieldInterceptor) GetOptions() map[string]any {
	return map[string]any{"symbol": i.symbol, "method": i.method}
}

func (i *useFunctionFieldInterceptor) Init(parser Parser, logger *slog.Logger) {
	cf := parseConverterFunctionConfigFromString(i.symbol)
	fn, ok := parser.FindFunction(cf.PackagePath, cf.TypeName)
	if ok {
		if len(fn.Params) == 1 && len(fn.Results) == 1 {
			i.function = &funcConverter{
				sourceType: fn.Params[0],
				targetType: fn.Results[0],
				pkgPath:    fn.PackagePath,
				funcName:   fn.Name,
			}
		}
	}

	if i.function == nil {
		varFns := parser.FindVariableMethods(cf.PackagePath, cf.TypeName)
		if len(varFns) > 0 {
			variableName := cf.TypeName
			for _, vfn := range varFns {
				if vfn.Name != i.method {
					continue
				}

				if len(vfn.Params) == 1 && len(vfn.Results) == 1 {
					i.function = &funcConverter{
						sourceType:   vfn.Params[0],
						targetType:   vfn.Results[0],
						variableName: &variableName,
						pkgPath:      vfn.PackagePath,
						funcName:     vfn.Name,
					}
				}
			}
		}
	}

	if i.function == nil {
		logger.Warn(util.ColorYellow(fmt.Sprintf("\tthere is no function matched with symbol=%s, method=%s", i.symbol, i.method)))
	}
}

func (i *useFunctionFieldInterceptor) canConvert(targetType, sourceType types.Type) bool {
	if i.function == nil {
		return false
	}

	if !TypeUtil.IsIdentical(targetType, i.function.targetType) {
		return false
	}

	if !TypeUtil.IsIdentical(sourceType, i.function.sourceType) {
		return false
	}
	return true
}

func (i *useFunctionFieldInterceptor) InterceptCanConvert(converter Converter, ctx LookupContext, targetType, sourceType types.Type) bool {
	return i.canConvert(targetType, sourceType) || converter.CanConvert(ctx, targetType, sourceType)
}

func (i *useFunctionFieldInterceptor) InterceptConvertField(converter Converter, ctx ConverterContext, target, source Symbol) jen.Code {
	if !i.canConvert(target.Type, source.Type) {
		return converter.ConvertField(ctx, target, source)
	}

	code := target.Expr().Op("=")
	if i.function.variableName != nil {
		return code.Qual(i.function.pkgPath, *i.function.variableName).Dot(i.function.funcName).Params(source.Expr())
	}
	return code.Qual(i.function.pkgPath, i.function.funcName).Params(source.Expr())
}

var _ FieldInterceptor = (*useFunctionFieldInterceptor)(nil)
